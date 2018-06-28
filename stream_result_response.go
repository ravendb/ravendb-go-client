package ravendb

import (
	"io"
	"net/http"
)

type StreamResultResponse struct {
	response *http.Response `json:"Response"`
	stream   io.Reader      `json:"Stream"`
}

func NewStreamResultResponse() *StreamResultResponse {
	return &StreamResultResponse{}
}

func (r *StreamResultResponse) getResponse() *http.Response {
	return r.response
}

func (r *StreamResultResponse) setResponse(response *http.Response) {
	r.response = response
}

func (r *StreamResultResponse) getStream() io.Reader {
	return r.stream
}

func (r *StreamResultResponse) setStream(stream io.Reader) {
	r.stream = stream
}
