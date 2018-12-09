package ravendb

import (
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/kjk/inflect"
)

var (
	// Note: helps find places that in Java code used DocumentConventsion.GetIdentityProperty()
	// if we add support for that
	DocumentConventions_identityPropertyName = "ID"
)

type DocumentIDGeneratorFunc func(dbName string, entity interface{}) string

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

	_maxHttpCacheSize int
	mu                sync.Mutex
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
		MaxLengthOfQueryUsingGetURL:                     1024 + 512,
		IdentityPartsSeparator:                          "/",
		_disableTopologyUpdates:                         false,
		RaiseIfQueryPageSizeIsNotSet:                    false,
		_transformClassCollectionNameToDocumentIdPrefix: DocumentConventions_defaultTransformCollectionNameToDocumentIdPrefix,
		_maxNumberOfRequestsPerSession:                  32,
		_maxHttpCacheSize:                               128 * 1024 * 1024,
	}
}

func (c *DocumentConventions) getMaxHttpCacheSize() int {
	return c._maxHttpCacheSize
}

func (c *DocumentConventions) Freeze() {
	c._frozen = true
}

func (c *DocumentConventions) GetCollectionName(entityOrClazz interface{}) string {
	return DefaultGetCollectionName(entityOrClazz)
}

func (c *DocumentConventions) IsThrowIfQueryPageSizeIsNotSet() bool {
	return c._throwIfQueryPageSizeIsNotSet
}

func DefaultGetCollectionName(entityOrClazz interface{}) string {
	// TODO: caching
	name := GetShortTypeNameName(entityOrClazz)
	return inflect.ToPlural(name)
}

func (c *DocumentConventions) UpdateFrom(configuration *ClientConfiguration) {
	if configuration == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if configuration.IsDisabled && c._originalConfiguration == nil {
		// nothing to do
		return
	}

	if configuration.IsDisabled && c._originalConfiguration != nil {
		// need to revert to original values
		c._maxNumberOfRequestsPerSession = c._originalConfiguration.MaxNumberOfRequestsPerSession
		c._readBalanceBehavior = c._originalConfiguration.ReadBalanceBehavior

		c._originalConfiguration = nil
		return
	}

	if c._originalConfiguration == nil {
		c._originalConfiguration = &ClientConfiguration{}
		c._originalConfiguration.Etag = -1
		c._originalConfiguration.MaxNumberOfRequestsPerSession = c._maxNumberOfRequestsPerSession
		c._originalConfiguration.ReadBalanceBehavior = c._readBalanceBehavior
	}

	c._maxNumberOfRequestsPerSession = firstNonZero(configuration.MaxNumberOfRequestsPerSession, c._originalConfiguration.MaxNumberOfRequestsPerSession)

	c._readBalanceBehavior = firstNonEmptyString(configuration.ReadBalanceBehavior, c._originalConfiguration.ReadBalanceBehavior)
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

func (c *DocumentConventions) GetReadBalanceBehavior() ReadBalanceBehavior {
	return c._readBalanceBehavior
}

func (c *DocumentConventions) Clone() *DocumentConventions {
	res := *c
	// mutex carries its locking state so we need to re-initialize it
	res.mu = sync.Mutex{}
	return &res
}

func (c *DocumentConventions) GetGoTypeName(entity interface{}) string {
	return GetFullTypeName(entity)
}

// returns "" if no identity property
func (c *DocumentConventions) GetIdentityProperty(clazz reflect.Type) string {
	return GetIdentityProperty(clazz)
}

func (c *DocumentConventions) GetDocumentIdGenerator() DocumentIDGeneratorFunc {
	return c._documentIdGenerator
}

func (c *DocumentConventions) SetDocumentIdGenerator(documentIdGenerator DocumentIDGeneratorFunc) {
	c._documentIdGenerator = documentIdGenerator
}

// Generates the document id.
func (c *DocumentConventions) GenerateDocumentId(databaseName string, entity interface{}) string {
	return c._documentIdGenerator(databaseName, entity)
}

func (c *DocumentConventions) IsDisableTopologyUpdates() bool {
	return c._disableTopologyUpdates
}

func (c *DocumentConventions) SetDisableTopologyUpdates(disable bool) {
	c._disableTopologyUpdates = disable
}

func (c *DocumentConventions) GetIdentityPartsSeparator() string {
	return c.IdentityPartsSeparator
}

func (c *DocumentConventions) GetTransformClassCollectionNameToDocumentIdPrefix() func(string) string {
	return c._transformClassCollectionNameToDocumentIdPrefix
}

func (c *DocumentConventions) DeserializeEntityFromJson(documentType reflect.Type, document TreeNode) (interface{}, error) {
	res, e := treeToValue(documentType, document)
	if e != nil {
		return nil, NewRavenException("Cannot deserialize entity %s", e)
	}
	return res, nil
}

func (c *DocumentConventions) TryConvertValueForQuery(fieldName string, value interface{}, forRange bool, stringValue *string) bool {
	// TODO: implement me
	// Tested by CustomSerializationTest
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
