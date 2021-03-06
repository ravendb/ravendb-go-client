package ravendb

import "strings"

// getRequest represents get request
type getRequest struct {
	url     string
	headers map[string]string
	query   string
	method  string
	content IContent
}

func (r *getRequest) getUrlAndQuery() string {
	if r.query == "" {
		return r.url
	}

	if strings.HasPrefix(r.query, "?") {
		return r.url + r.query
	}

	return r.url + "?" + r.query
}

type IContent interface {
	writeContent() map[string]interface{}
}
