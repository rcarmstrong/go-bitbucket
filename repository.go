package bitbucket

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/mitchellh/mapstructure"
)

type Project struct {
	Key  string
	Name string
}

type Repository struct {
	c *Client

	Project     Project
	Slug        string
	Full_name   string
	Description string
	Fork_policy string
	Type        string
	Owner       map[string]interface{}
	Links       map[string]interface{}
}

type Pipeline struct {
	Type       string
	Enabled    bool
	Repository Repository
}

type PipelineVariable struct {
	Type    string
	Uuid    string
	Key     string
	Value   string
	Secured bool
}

type PipelineKeyPair struct {
	Type       string
	Uuid       string
	PublicKey  string
	PrivateKey string
}

func (r *Repository) Create(ro *RepositoryOptions) (*Repository, error) {
	data := r.buildRepositoryBody(ro)
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.Repo_slug)
	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) Get(ro *RepositoryOptions) (*Repository, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.Repo_slug)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) Delete(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.Repo_slug)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) ListWatchers(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/watchers", ro.Owner, ro.Repo_slug)
	return r.c.execute("GET", urlStr, "")
}

func (r *Repository) ListForks(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/forks", ro.Owner, ro.Repo_slug)
	return r.c.execute("GET", urlStr, "")
}

// ListDefaultReviewers returns the list of default reviewers for the given repo
func (r *Repository) ListDefaultReviewers(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers", ro.Owner, ro.Repo_slug)
	return r.c.execute(http.MethodGet, urlStr, "")
}

// AddDefaultReviewer will add the given user to the default-reviewers list. The RepositoryOptions for the Owner
// and the Repo_slug are used. The username is not validated. Review for spelling mistakes.
func (r *Repository) AddDefaultReviewer(ro *RepositoryOptions, username string) error {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers/%s", ro.Owner, ro.Repo_slug, username)
	_, err := r.c.execute(http.MethodPut, urlStr, "")
	if err != nil {
		return err
	}

	return nil
}

// RemoveDefaultReviewer will take the given user out of the default-reviewers list. The RepositoryOptions for the Owner
// and the Repo_slug are used. The username is not validated. Review for spelling mistakes.
func (r *Repository) RemoveDefaultReviewer(ro *RepositoryOptions, username string) error {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers/%s", ro.Owner, ro.Repo_slug, username)
	_, err := r.c.execute(http.MethodDelete, urlStr, "")
	if err != nil {
		return err
	}

	return nil
}

// UploadFile takes in the full path of the desired file, ie /src/main/test.txt, and uses the content string to
// create the file in the specified repo. The Owner and Repo_slug fields are needed from the RepositoryOptions.
func (r *Repository) UploadFile(ro *RepositoryOptions, branch, filePath, content string) (*http.Response, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/src", ro.Owner, ro.Repo_slug)
	client := http.DefaultClient

	data := url.Values{}
	data.Set(filePath, content)
	data.Add("branch", branch)

	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	if r.c.Auth.user != "" && r.c.Auth.password != "" {
		req.SetBasicAuth(r.c.Auth.user, r.c.Auth.password)
	} else if r.c.Auth.token.Valid() {
		r.c.Auth.token.SetAuthHeader(req)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	// Submit the request
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (r *Repository) GetFile(ro *RepositoryOptions, filePath, hash string) ([]byte, error) {
	if hash == "" {
		hash = "master"
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/src/%s/%s", ro.Owner, ro.Repo_slug, hash, filePath)

	return r.c.executeRaw("GET", urlStr, "")
}

func (r *Repository) UpdatePipelineConfig(rpo *RepositoryPipelineOptions) (*Pipeline, error) {
	data := r.buildPipelineBody(rpo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config", rpo.Owner, rpo.Repo_slug)
	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineRepository(response)
}

func (r *Repository) AddPipelineVariable(rpvo *RepositoryPipelineVariableOptions) (*PipelineVariable, error) {
	data := r.buildPipelineVariableBody(rpvo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/", rpvo.Owner, rpvo.Repo_slug)

	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineVariableRepository(response)
}

func (r *Repository) AddPipelineKeyPair(rpkpo *RepositoryPipelineKeyPairOptions) (*PipelineKeyPair, error) {
	data := r.buildPipelineKeyPairBody(rpkpo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/ssh/key_pair", rpkpo.Owner, rpkpo.Repo_slug)

	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineKeyPairRepository(response)
}

func (r *Repository) buildJsonBody(body map[string]interface{}) string {

	data, err := json.Marshal(body)
	if err != nil {
		pp.Println(err)
		os.Exit(9)
	}

	return string(data)
}

func (r *Repository) buildRepositoryBody(ro *RepositoryOptions) string {

	body := map[string]interface{}{}

	if ro.Scm != "" {
		body["scm"] = ro.Scm
	}
	//if ro.Scm != "" {
	//		body["name"] = ro.Name
	//}
	if ro.Is_private != "" {
		body["is_private"] = ro.Is_private
	}
	if ro.Description != "" {
		body["description"] = ro.Description
	}
	if ro.Fork_policy != "" {
		body["fork_policy"] = ro.Fork_policy
	}
	if ro.Language != "" {
		body["language"] = ro.Language
	}
	if ro.Has_issues != "" {
		body["has_issues"] = ro.Has_issues
	}
	if ro.Has_wiki != "" {
		body["has_wiki"] = ro.Has_wiki
	}
	if ro.Project != "" {
		body["project"] = map[string]string{
			"key": ro.Project,
		}
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineBody(rpo *RepositoryPipelineOptions) string {

	body := map[string]interface{}{}

	body["enabled"] = rpo.Enabled

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineVariableBody(rpvo *RepositoryPipelineVariableOptions) string {

	body := map[string]interface{}{}

	if rpvo.Uuid != "" {
		body["uuid"] = rpvo.Uuid
	}
	body["key"] = rpvo.Key
	body["value"] = rpvo.Value
	body["secured"] = rpvo.Secured

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineKeyPairBody(rpkpo *RepositoryPipelineKeyPairOptions) string {

	body := map[string]interface{}{}

	if rpkpo.Private_key != "" {
		body["private_key"] = rpkpo.Private_key
	}
	if rpkpo.Public_key != "" {
		body["public_key"] = rpkpo.Public_key
	}

	return r.buildJsonBody(body)
}

func decodeRepository(repoResponse interface{}) (*Repository, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var repository = new(Repository)
	err := mapstructure.Decode(repoMap, repository)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func decodePipelineRepository(repoResponse interface{}) (*Pipeline, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipeline = new(Pipeline)
	err := mapstructure.Decode(repoMap, pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

func decodePipelineVariableRepository(repoResponse interface{}) (*PipelineVariable, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineVariable = new(PipelineVariable)
	err := mapstructure.Decode(repoMap, pipelineVariable)
	if err != nil {
		return nil, err
	}

	return pipelineVariable, nil
}

func decodePipelineKeyPairRepository(repoResponse interface{}) (*PipelineKeyPair, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineKeyPair = new(PipelineKeyPair)
	err := mapstructure.Decode(repoMap, pipelineKeyPair)
	if err != nil {
		return nil, err
	}

	return pipelineKeyPair, nil
}
