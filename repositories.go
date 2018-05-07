package bitbucket

import (
	"fmt"
	"net/url"
)

//"github.com/k0kubun/pp"

type Repositories struct {
	c                  *Client
	PullRequests       *PullRequests
	Repository         *Repository
	Commits            *Commits
	Diff               *Diff
	BranchRestrictions *BranchRestrictions
	Webhooks           *Webhooks
	repositories
}

func (r *Repositories) ListForAccount(ro *RepositoriesOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s", ro.Owner)
	if ro.Role != "" {
		urlStr += "?role=" + ro.Role
	}
	return r.c.execute("GET", urlStr, "")
}

func (r *Repositories) ListForTeam(ro *RepositoriesOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s", ro.Owner)
	if ro.Role != "" {
		urlStr += "?role=" + ro.Role
	}
	return r.c.execute("GET", urlStr, "")
}

func (r *Repositories) ListPublic() (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/", "")
	return r.c.execute("GET", urlStr, "")
}

// ListForProject returns a pagenated list of repositories for the given project
func (r *Repositories) ListForProject(ro *ProjectRepositoryOptions) (interface{}, error) {
	values, _ := url.ParseQuery(fmt.Sprintf("q=project.key=\"%s\"&pagelen=%d&page=%d", ro.Project, ro.PageLength, ro.Page))
	urlStr := r.c.requestUrl("/repositories/%s?%s", ro.Owner, values.Encode())
	return r.c.execute("GET", urlStr, "")
}
