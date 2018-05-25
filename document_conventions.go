package ravendb

import "time"

type DocumentIDGeneratorFunc func(dbName string, entity Object) string

// DocumentConventions describes document conventions
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/data/document_conventions.py#L9
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/conventions/DocumentConventions.java#L31
type DocumentConventions struct {
	MaxNumberOfRequestsPerSession int
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

	_documentIdGenerator DocumentIDGeneratorFunc
}

// NewDocumentConventions creates DocumentConventions with default values
func NewDocumentConventions() *DocumentConventions {
	return &DocumentConventions{
		MaxNumberOfRequestsPerSession: 32,
		MaxLengthOfQueryUsingGetURL:   1024 + 512,
		IdentityPartsSeparator:        "/",
		DisableTopologyUpdate:         false,
		RaiseIfQueryPageSizeIsNotSet:  false,
	}
}

func (c *DocumentConventions) getCollectionName(entity Object) string {
	return defaultGetCollectionName(entity)
}

// TODO: tests
func defaultGetCollectionName(entity interface{}) string {
	// TODO: caching
	typ := getShortTypeName(entity)
	result := pluralize(typ)
	return result
}

func (c *DocumentConventions) getGoTypeName(entity interface{}) string {
	return getFullTypeName(entity)
}

func (c *DocumentConventions) getDocumentIdGenerator() DocumentIDGeneratorFunc {
	return c._documentIdGenerator
}

func (c *DocumentConventions) setDocumentIdGenerator(documentIdGenerator DocumentIDGeneratorFunc) {
	c._documentIdGenerator = documentIdGenerator
}

// Generates the document id.
func (c *DocumentConventions) generateDocumentId(databaseName String, entity Object) String {
	return c._documentIdGenerator(databaseName, entity)
}
