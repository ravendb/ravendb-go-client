package ravendb

import "net/http"

type GetResponse struct {
	result       []byte
	headers      map[string]string
	statusCode   int
	isForceRetry bool
}

func NewGetResponse() *GetResponse {
	return &GetResponse{
		headers: map[string]string{},
	}
}

func (r *GetResponse) requestHasErrors() bool {
	switch r.statusCode {
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
