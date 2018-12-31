package ravendb

import (
	"net/http"
	"strings"
)

func gttpExtensionsGetRequiredEtagHeader(response *http.Response) (*string, error) {
	hdr := response.Header.Get(headersEtag)
	if hdr == "" {
		return nil, newIllegalStateError("Response did't had an ETag header")
	}
	etag := httpExtensionsEtagHeaderToChangeVector(hdr)
	return &etag, nil
}

func gttpExtensionsGetEtagHeader(response *http.Response) *string {
	hdr := response.Header.Get(headersEtag)
	if hdr == "" {
		return nil
	}
	res := httpExtensionsEtagHeaderToChangeVector(hdr)
	return &res
}

func gttpExtensionsGetEtagHeaderFromMap(headers map[string]string) *string {
	hdr := headers[headersEtag]
	if hdr == "" {
		return nil
	}
	res := httpExtensionsEtagHeaderToChangeVector(hdr)
	return &res
}

// TODO: add test
func httpExtensionsEtagHeaderToChangeVector(responseHeader string) string {
	panicIf(responseHeader == "", "Response did't had an ETag header")

	if strings.HasPrefix(responseHeader, `"`) {
		return responseHeader[1 : len(responseHeader)-1]
	}

	return responseHeader
}

func httpExtensionsGetBooleanHeader(response *http.Response, header string) bool {
	hdr := response.Header.Get(header)
	return strings.EqualFold(hdr, "true")
}
