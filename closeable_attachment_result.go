package ravendb

import (
	"io"
	"net/http"
)

// Note: In Java it's CloseableAttachmentResult

// AttachmentResult represents an attachment
type AttachmentResult struct {
	Data     io.Reader
	Details  *AttachmentDetails
	response *http.Response
}

func newAttachmentResult(response *http.Response, details *AttachmentDetails) *AttachmentResult {
	return &AttachmentResult{
		Data:     response.Body,
		Details:  details,
		response: response,
	}
}

// Close closes the attachment
func (r *AttachmentResult) Close() error {
	if r.response.Body != nil {
		return r.response.Body.Close()
	}
	return nil
}
