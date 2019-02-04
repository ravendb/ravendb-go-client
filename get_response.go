package ravendb

import "net/http"

// GetResponse represents result of get request
type GetResponse struct {
	Result       []byte
	Headers      map[string]string
	StatusCode   int
	IsForceRetry bool
}

func (r *GetResponse) requestHasErrors() bool {
	switch r.StatusCode {
	case 0,
		http.StatusOK,
		http.StatusCreated,
		http.StatusNoContent,
		http.StatusNotModified,
		http.StatusNonAuthoritativeInfo,
		http.StatusNotFound:
		return false
	default:
		return true
	}
}
