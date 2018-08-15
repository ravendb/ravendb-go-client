package ravendb

import (
	"io"
	"net/http"
)

type CloseableAttachmentResult struct {
	details  *AttachmentDetails
	response *http.Response
}

func NewCloseableAttachmentResult(response *http.Response, details *AttachmentDetails) *CloseableAttachmentResult {
	return &CloseableAttachmentResult{
		details:  details,
		response: response,
	}
}

func (r *CloseableAttachmentResult) GetData() io.Reader {
	return r.response.Body
}

func (r *CloseableAttachmentResult) getDetails() *AttachmentDetails {
	return r.details
}

func (r *CloseableAttachmentResult) Close() {
	if r.response.Body != nil {
		r.response.Body.Close()
	}
}
