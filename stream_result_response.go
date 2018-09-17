package ravendb

import (
	"io"
	"net/http"
)

type StreamResultResponse struct {
	Response *http.Response `json:"Response"`
	Stream   io.Reader      `json:"Stream"`
}
