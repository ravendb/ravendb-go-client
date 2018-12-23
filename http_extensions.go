package ravendb

import (
	"net/http"
	"strings"
)

func HttpExtensions_getRequiredEtagHeader(response *http.Response) (*string, error) {
	hdr := response.Header.Get(headersEtag)
	if hdr == "" {
		return nil, newIllegalStateError("Response did't had an ETag header")
	}
	etag := HttpExtensions_etagHeaderToChangeVector(hdr)
	return &etag, nil
}

func HttpExtensions_getEtagHeader(response *http.Response) *string {
	hdr := response.Header.Get(headersEtag)
	if hdr == "" {
		return nil
	}
	res := HttpExtensions_etagHeaderToChangeVector(hdr)
	return &res
}

func HttpExtensions_getEtagHeaderFromMap(headers map[string]string) *string {
	hdr := headers[headersEtag]
	if hdr == "" {
		return nil
	}
	res := HttpExtensions_etagHeaderToChangeVector(hdr)
	return &res
}

// TODO: add test
func HttpExtensions_etagHeaderToChangeVector(responseHeader string) string {
	panicIf(responseHeader == "", "Response did't had an ETag header")

	if strings.HasPrefix(responseHeader, `"`) {
		return responseHeader[1 : len(responseHeader)-1]
	}

	return responseHeader
}

func HttpExtensions_getBooleanHeader(response *http.Response, header string) bool {
	hdr := response.Header.Get(header)
	return strings.EqualFold(hdr, "true")
}
