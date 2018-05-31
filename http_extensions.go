package ravendb

import (
	"net/http"
	"strings"
)

func HttpExtensions_getEtagHeader(response *http.Response) *string {
	hdr := response.Header.Get(Constants_Headers_ETAG)
	if hdr == "" {
		return nil
	}
	res := HttpExtensions_etagHeaderToChangeVector(hdr)
	return &res
}

// TODO: add test
func HttpExtensions_etagHeaderToChangeVector(responseHeader String) String {
	panicIf(responseHeader == "", "Response did't had an ETag header")

	if strings.HasPrefix(responseHeader, `"`) {
		return responseHeader[1 : len(responseHeader)-1]
	}

	return responseHeader
}
