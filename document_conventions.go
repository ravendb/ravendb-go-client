package ravendb

import (
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	// Note: helps find places that in Java code used DocumentConventsion.getIdentityProperty()
	// if we add support for that
	DocumentConventions_identityPropertyName = "ID"
)

type DocumentIDGeneratorFunc func(dbName string, entity Object) string

// DocumentConventions describes document conventions
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

	_throwIfQueryPageSizeIsNotSet bool

	mu sync.Mutex
}

// Note: Java has it as frozen global variable (possibly for perf) but Go
// has no notion of frozen objects so for safety we create new object
// (avoids accidental modification of shared, global state)
// TODO: replace with direct calls to NewDocumentConventions()
func DocumentConventions_defaultConventions() *DocumentConventions {
	return NewDocumentConventions()
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

func (c *DocumentConventions) getCollectionName(entityOrClazz Object) string {
	return defaultGetCollectionName(entityOrClazz)
}

func (c *DocumentConventions) isThrowIfQueryPageSizeIsNotSet() bool {
	return c._throwIfQueryPageSizeIsNotSet
}

func defaultGetCollectionName(entityOrClazz interface{}) string {
	// TODO: caching
	name := getShortTypeName(entityOrClazz)
	return pluralize(name)
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

func DocumentConventions_defaultTransformCollectionNameToDocumentIdPrefix(collectionName string) string {
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
func (c *DocumentConventions) generateDocumentId(databaseName string, entity Object) string {
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

func (c *DocumentConventions) deserializeEntityFromJson(documentType reflect.Type, document ObjectNode) (interface{}, error) {
	res, e := treeToValue(documentType, document)
	if e != nil {
		return nil, NewRavenException("Cannot deserialize entity %s", e)
	}
	return res, nil
}

func (c *DocumentConventions) tryConvertValueForQuery(fieldName string, value Object, forRange bool, stringValue *string) bool {
	panicIf(true, "NYI")
	/*
		for (Tuple<Class, IValueForQueryConverter<Object>> queryValueConverter : _listOfQueryValueConverters) {
			if (!queryValueConverter.first.isInstance(value)) {
				continue;
			}

			return queryValueConverter.second.tryConvertValueForQuery(fieldName, value, forRange, stringValue);
		}
	*/
	*stringValue = ""
	return false
}
