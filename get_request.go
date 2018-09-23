package ravendb

import "strings"

type GetRequest struct {
	url     string
	headers map[string]string
	query   string
	method  string
	content IContent
}

func (r *GetRequest) getUrlAndQuery() string {
	if r.query == "" {
		return r.url
	}

	if strings.HasPrefix(r.query, "?") {
		return r.url + r.query
	}

	return r.url + "?" + r.query
}

func (r *GetRequest) GetRequest() {
	r.headers = make(map[string]string)
}

type IContent interface {
	writeContent() map[string]interface{}
}
