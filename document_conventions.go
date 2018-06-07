package ravendb

import (
	"strings"
	"sync"
	"time"
	"unicode"
)

type DocumentIDGeneratorFunc func(dbName string, entity Object) string

// DocumentConventions describes document conventions
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/data/document_conventions.py#L9
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/conventions/DocumentConventions.java#L31
type DocumentConventions struct {
	_frozen                bool
	_originalConfiguration *ClientConfiguration

	_maxNumberOfRequestsPerSession int
	// timeout for wait to server
	Timeout                  time.Duration
	UseOptimisticConcurrency bool
	// JsonDdefaultMethod = DocumentConventions.json_default
	MaxLengthOfQueryUsingGetURL int
	IdentityPartsSeparator      string
	_disableTopologyUpdates     bool
	// If set to 'true' then it will return an error when any query is performed (in session)
	// without explicit page size set
	RaiseIfQueryPageSizeIsNotSet bool // TODO: rename to ErrorIfQueryPageSizeIsNotSet

	_documentIdGenerator DocumentIDGeneratorFunc

	_readBalanceBehavior                            ReadBalanceBehavior
	_transformClassCollectionNameToDocumentIdPrefix func(string) string
	mu                                              sync.Mutex
}

var (
	DocumentConventions_defaultConventions *DocumentConventions
)

func init() {
	DocumentConventions_defaultConventions = NewDocumentConventions()
	DocumentConventions_defaultConventions.freeze()
}

// NewDocumentConventions creates DocumentConventions with default values
func NewDocumentConventions() *DocumentConventions {
	return &DocumentConventions{
		_readBalanceBehavior:                            ReadBalanceBehavior_NONE,
		_maxNumberOfRequestsPerSession:                  32,
		MaxLengthOfQueryUsingGetURL:                     1024 + 512,
		IdentityPartsSeparator:                          "/",
		_disableTopologyUpdates:                         false,
		RaiseIfQueryPageSizeIsNotSet:                    false,
		_transformClassCollectionNameToDocumentIdPrefix: DocumentConventions_defaultTransformCollectionNameToDocumentIdPrefix,
	}

}

func (c *DocumentConventions) freeze() {
	c._frozen = true
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

func (c *DocumentConventions) updateFrom(configuration *ClientConfiguration) {
	if configuration == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if configuration.isDisabled() && c._originalConfiguration == nil {
		// nothing to do
		return
	}

	if configuration.isDisabled() && c._originalConfiguration != nil {
		// need to revert to original values
		c._maxNumberOfRequestsPerSession = c._originalConfiguration.getMaxNumberOfRequestsPerSession()
		c._readBalanceBehavior = c._originalConfiguration.getReadBalanceBehavior()

		c._originalConfiguration = nil
		return
	}

	if c._originalConfiguration == nil {
		c._originalConfiguration = NewClientConfiguration()
		c._originalConfiguration.setEtag(-1)
		c._originalConfiguration.setMaxNumberOfRequestsPerSession(c._maxNumberOfRequestsPerSession)
		c._originalConfiguration.setReadBalanceBehavior(c._readBalanceBehavior)
	}

	c._maxNumberOfRequestsPerSession = firstNonZero(configuration.getMaxNumberOfRequestsPerSession(), c._originalConfiguration.getMaxNumberOfRequestsPerSession())

	c._readBalanceBehavior = firstNonEmptyString(configuration.getReadBalanceBehavior(), c._originalConfiguration.getReadBalanceBehavior())
}

func DocumentConventions_defaultTransformCollectionNameToDocumentIdPrefix(collectionName String) String {
	upperCount := 0
	for _, c := range collectionName {
		if unicode.IsUpper(c) {
			upperCount++

			// multiple capital letters, so probably something that we want to preserve caps on.
			if upperCount > 1 {
				return collectionName
			}
		}
	}
	if upperCount == 1 {
		return strings.ToLower(collectionName)
	}
	// upperCount must be 0
	return collectionName
}

func (c *DocumentConventions) getReadBalanceBehavior() ReadBalanceBehavior {
	return c._readBalanceBehavior
}

func (c *DocumentConventions) clone() *DocumentConventions {
	res := *c
	return &res
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

func (c *DocumentConventions) isDisableTopologyUpdates() bool {
	return c._disableTopologyUpdates
}

func (c *DocumentConventions) setDisableTopologyUpdates(disable bool) {
	c._disableTopologyUpdates = disable
}

func (c *DocumentConventions) getIdentityPartsSeparator() string {
	return c.IdentityPartsSeparator
}

func (c *DocumentConventions) getTransformClassCollectionNameToDocumentIdPrefix() func(string) string {
	return c._transformClassCollectionNameToDocumentIdPrefix
}
