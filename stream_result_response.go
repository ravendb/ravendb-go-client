package ravendb

import (
	"io"
	"net/http"
)

type StreamResultResponse struct {
	Response *http.Response `json:"Response"`
	Stream   io.Reader      `json:"Stream"`
}

func NewStreamResultResponse() *StreamResultResponse {
	return &StreamResultResponse{}
}

func (r *StreamResultResponse) getResponse() *http.Response {
	return r.Response
}

func (r *StreamResultResponse) SetResponse(response *http.Response) {
	r.Response = response
}

func (r *StreamResultResponse) getStream() io.Reader {
	return r.Stream
}

func (r *StreamResultResponse) setStream(stream io.Reader) {
	r.Stream = stream
}
