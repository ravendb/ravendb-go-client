package ravendb

import "net/http"

func HttpExtensions_getEtagHeader(response *http.Response) *string {
	hdr := response.Header.Get(Constants_Headers_ETAG)
	if hdr == "" {
		return nil
	}
	return &HttpExtensions_etagHeaderToChangeVector(hdr)
}

func HttpExtensions_etagHeaderToChangeVector(responseHeader String) String {
	panicIf(responseHeader == "", "Response did't had an ETag header")

	if responseHeader.HasPrefix(`"`) {
		return responseHeader[1 : len(responseHeader)-1]
	}

	return responseHeader
}
