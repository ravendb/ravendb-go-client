package ravendb

import (
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	// Note: helps find places that in Java code used DocumentConventsion.GetIdentityProperty()
	// if we add support for that
	documentConventionsIdentityPropertyName = "ID"
)

type DocumentIDGeneratorFunc func(dbName string, entity interface{}) (string, error)

// DocumentConventions describes document conventions
type DocumentConventions struct {
	frozen                bool
	originalConfiguration *ClientConfiguration

	MaxNumberOfRequestsPerSession int
	// timeout for wait to server
	Timeout                  time.Duration
	UseOptimisticConcurrency bool
	// JsonDefaultMethod = DocumentConventions.json_default
	MaxLengthOfQueryUsingGetURL int
	IdentityPartsSeparator      string
	disableTopologyUpdates      bool
	// If set to 'true' then it will return an error when any query is performed (in session)
	// without explicit page size set
	RaiseIfQueryPageSizeIsNotSet bool // TODO: rename to ErrorIfQueryPageSizeIsNotSet

	documentIDGenerator DocumentIDGeneratorFunc

	// allows overriding entity -> collection name logic
	FindCollectionName func(interface{}) string

	ReadBalanceBehavior                            ReadBalanceBehavior
	transformClassCollectionNameToDocumentIDPrefix func(string) string

	// if true, will return error if page size is not set
	ErrorIfQueryPageSizeIsNotSet bool

	maxHttpCacheSize int

	// a pointer to silence go vet when copying DocumentConventions wholesale
	mu *sync.Mutex
}

// Note: Java has it as frozen global variable (possibly for perf) but Go
// has no notion of frozen objects so for safety we create new object
// (avoids accidental modification of shared, global state)
// TODO: replace with direct calls to NewDocumentConventions()
func getDefaultConventions() *DocumentConventions {
	return NewDocumentConventions()
}

// NewDocumentConventions creates DocumentConventions with default values
func NewDocumentConventions() *DocumentConventions {
	return &DocumentConventions{
		ReadBalanceBehavior:                            ReadBalanceBehaviorNone,
		MaxLengthOfQueryUsingGetURL:                    1024 + 512,
		IdentityPartsSeparator:                         "/",
		disableTopologyUpdates:                         false,
		RaiseIfQueryPageSizeIsNotSet:                   false,
		transformClassCollectionNameToDocumentIDPrefix: getDefaultTransformCollectionNameToDocumentIdPrefix,
		MaxNumberOfRequestsPerSession:                  32,
		maxHttpCacheSize:                               128 * 1024 * 1024,
		mu:                                             &sync.Mutex{},
	}
}

func (c *DocumentConventions) getMaxHttpCacheSize() int {
	return c.maxHttpCacheSize
}

func (c *DocumentConventions) Freeze() {
	c.frozen = true
}

// GetCollectionNameDefault is a default way of
func GetCollectionNameDefault(entityOrType interface{}) string {
	name := getShortTypeNameForEntityOrType(entityOrType)
	return ToPlural(name)
}

func (c *DocumentConventions) getCollectionName(entityOrType interface{}) string {
	if c.FindCollectionName != nil {
		return c.FindCollectionName(entityOrType)
	}
	return GetCollectionNameDefault(entityOrType)
}

func getCollectionNameForTypeOrEntity(entityOrType interface{}) string {
	name := getShortTypeNameForEntityOrType(entityOrType)
	return ToPlural(name)
}

func (c *DocumentConventions) UpdateFrom(configuration *ClientConfiguration) {
	if configuration == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if configuration.IsDisabled && c.originalConfiguration == nil {
		// nothing to do
		return
	}

	if configuration.IsDisabled && c.originalConfiguration != nil {
		// need to revert to original values
		c.MaxNumberOfRequestsPerSession = c.originalConfiguration.MaxNumberOfRequestsPerSession
		c.ReadBalanceBehavior = c.originalConfiguration.ReadBalanceBehavior

		c.originalConfiguration = nil
		return
	}

	if c.originalConfiguration == nil {
		c.originalConfiguration = &ClientConfiguration{}
		c.originalConfiguration.Etag = -1
		c.originalConfiguration.MaxNumberOfRequestsPerSession = c.MaxNumberOfRequestsPerSession
		c.originalConfiguration.ReadBalanceBehavior = c.ReadBalanceBehavior
	}

	c.MaxNumberOfRequestsPerSession = firstNonZero(configuration.MaxNumberOfRequestsPerSession, c.originalConfiguration.MaxNumberOfRequestsPerSession)

	c.ReadBalanceBehavior = firstNonEmptyString(configuration.ReadBalanceBehavior, c.originalConfiguration.ReadBalanceBehavior)
}

func getDefaultTransformCollectionNameToDocumentIdPrefix(collectionName string) string {
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

func (c *DocumentConventions) Clone() *DocumentConventions {
	res := *c
	// mutex carries its locking state so we need to re-initialize it
	res.mu = &sync.Mutex{}
	return &res
}

func (c *DocumentConventions) getGoTypeName(entity interface{}) string {
	return getFullTypeName(entity)
}

// returns "" if no identity property
func (c *DocumentConventions) GetIdentityProperty(clazz reflect.Type) string {
	return getIdentityProperty(clazz)
}

func (c *DocumentConventions) GetDocumentIDGenerator() DocumentIDGeneratorFunc {
	return c.documentIDGenerator
}

func (c *DocumentConventions) SetDocumentIDGenerator(documentIDGenerator DocumentIDGeneratorFunc) {
	c.documentIDGenerator = documentIDGenerator
}

// Generates the document id.
func (c *DocumentConventions) GenerateDocumentID(databaseName string, entity interface{}) (string, error) {
	return c.documentIDGenerator(databaseName, entity)
}

func (c *DocumentConventions) IsDisableTopologyUpdates() bool {
	return c.disableTopologyUpdates
}

func (c *DocumentConventions) SetDisableTopologyUpdates(disable bool) {
	c.disableTopologyUpdates = disable
}

func (c *DocumentConventions) GetIdentityPartsSeparator() string {
	return c.IdentityPartsSeparator
}

func (c *DocumentConventions) GetTransformClassCollectionNameToDocumentIdPrefix() func(string) string {
	return c.transformClassCollectionNameToDocumentIDPrefix
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
