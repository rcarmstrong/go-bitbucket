package bitbucket

type Teams struct {
	c *Client
}

func (t *Teams) List(role string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/?role=%s", role)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Profile(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Members(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/members", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Followers(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/followers", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Following(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/following", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Repositories(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/repositories", teamname)
	return t.c.execute("GET", urlStr, "")
}

// Projects returns a list of project names for the given team.
func (t *Teams) Projects(teamname string) ([]string, error) {
	urlStr := t.c.requestUrl("/teams/%s/repositories", teamname)
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	list := response.(map[string]interface{})["values"].([]interface{})
	projects := make([]string, len(list))

	for i, l := range list {
		projects[i] = l.(map[string]interface{})["name"].(string)
	}

	return projects, nil
}
