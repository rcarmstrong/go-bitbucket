package bitbucket

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net/http"
	urls "net/url"
	"strconv"
	"strings"
)

type Client struct {
	Auth         *auth
	Users        users
	User         user
	Teams        teams
	Repositories *Repositories
	Pagelen      uint64
}

type auth struct {
	app_id, secret string
	user, password string
}

func NewOAuth(i, s string) *Client {
	a := &auth{app_id: i, secret: s}
	return injectClient(a)
}

func NewBasicAuth(u, p string) *Client {
	a := &auth{user: u, password: p}
	return injectClient(a)
}

const DEFAULT_PAGE_LENGHT = 10

func injectClient(a *auth) *Client {
	c := &Client{Auth: a, Pagelen: DEFAULT_PAGE_LENGHT}
	c.Repositories = &Repositories{
		c:                  c,
		PullRequests:       &PullRequests{c: c},
		Repository:         &Repository{c: c},
		Commits:            &Commits{c: c},
		Diff:               &Diff{c: c},
		BranchRestrictions: &BranchRestrictions{c: c},
		Webhooks:           &Webhooks{c: c},
	}
	c.Users = &Users{c: c}
	c.User = &User{c: c}
	c.Teams = &Teams{c: c}
	return c
}

func (c *Client) execute(method, url, text string) interface{} {

	// Use pagination if changed from default value
	const DEC_RADIX = 10
	if strings.Contains(url, "/repositories/") {
		if c.Pagelen != DEFAULT_PAGE_LENGHT {
			urlObj, err := urls.Parse(url)
			if err != nil {
				return err
			}
			q := urlObj.Query()
			q.Set("pagelen", strconv.FormatUint(c.Pagelen, DEC_RADIX))
			urlObj.RawQuery = q.Encode()
			url = urlObj.String()
		}
	}

	body := strings.NewReader(text)
	req, err := http.NewRequest(method, url, body)
	if text != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return err
	}

	if c.Auth.user != "" && c.Auth.password != "" {
		req.SetBasicAuth(c.Auth.user, c.Auth.password)
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result interface{}
	json.Unmarshal(buf, &result)

	return result
}

func (c *Client) requestUrl(template string, args ...interface{}) string {

	if len(args) == 1 && args[0] == "" {
		return GetApiBaseURL() + template
	}
	return GetApiBaseURL() + fmt.Sprintf(template, args...)
}
