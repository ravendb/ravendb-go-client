package ravendb

import "time"

// DocumentConventions describes document conventions
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/data/document_conventions.py#L9
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/conventions/DocumentConventions.java#L31
type DocumentConventions struct {
	MaxNumberOfRequestsPerSession int
	MaxIdsToCatch                 int
	// timeout for wait to server
	Timeout                  time.Duration
	UseOptimisticConcurrency bool
	// JsonDdefaultMethod = DocumentConventions.json_default
	MaxLengthOfQueryUsingGetURL int
	IdentityPartsSeparator      string
	DisableTopologyUpdate       bool
	// If set to 'true' then it will return an error when any query is performed (in session)
	// without explicit page size set
	RaiseIfQueryPageSizeIsNotSet bool // TODO: rename to ErrorIfQueryPageSizeIsNotSet
}

// NewDocumentConventions creates DocumentConventions with default values
func NewDocumentConventions() *DocumentConventions {
	return &DocumentConventions{
		MaxNumberOfRequestsPerSession: 32,
		MaxIdsToCatch:                 32,
		Timeout:                       time.Second * 30,
		MaxLengthOfQueryUsingGetURL:   1024 + 512,
		IdentityPartsSeparator:        "/",
		DisableTopologyUpdate:         false,
		RaiseIfQueryPageSizeIsNotSet:  false,
	}
}

func (c *DocumentConventions) getCollectionName(entity Object) string {
	// TODO: implement me
	panicIf(true, "NYI")
	return ""
}
