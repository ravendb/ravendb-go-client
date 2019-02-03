package ravendb

import (
	"log"
	"reflect"
)

type SubscriptionBatchItem struct {
	_result          interface{}
	exceptionMessage string
	ID               string
	changeVector     string

	rawResult   map[string]interface{}
	rawMetadata map[string]interface{}
	_metadata   *MetadataAsDictionary
}

func (i *SubscriptionBatchItem) throwItemProcessException() error {
	return newIllegalStateError("Failed to process document " + i.ID + " with Change Vector " + i.changeVector + " because: \n" + i.exceptionMessage)
}

func (i *SubscriptionBatchItem) GetResult() (interface{}, error) {
	if i.exceptionMessage != "" {
		return nil, i.throwItemProcessException()
	}
	return i._result, nil
}

func (i *SubscriptionBatchItem) GetMetadata() *MetadataAsDictionary {
	if i._metadata == nil {
		i._metadata = NewMetadataAsDictionary(i.rawMetadata, nil, "")
	}

	return i._metadata
}

type SubscriptionBatch struct {
	_clazz                       reflect.Type
	_revisions                   bool
	_requestExecutor             *RequestExecutor
	_store                       *DocumentStore
	_dbName                      string
	_logger                      *log.Logger
	_generateEntityIdOnTheClient *generateEntityIDOnTheClient

	Items []*SubscriptionBatchItem
}

func (b *SubscriptionBatch) getNumberOfItemsInBatch() int {
	return len(b.Items)
}

func (b *SubscriptionBatch) openSession() (*DocumentSession, error) {
	sessionOptions := &SessionOptions{
		Database:        b._dbName,
		RequestExecutor: b._requestExecutor,
	}
	return b._store.OpenSessionWithOptions(sessionOptions)
}

func NewSubscriptionBatch(clazz reflect.Type, revisions bool, requestExecutor *RequestExecutor, store *DocumentStore, dbName string, logger *log.Logger) *SubscriptionBatch {
	res := &SubscriptionBatch{
		_clazz:           clazz,
		_revisions:       revisions,
		_requestExecutor: requestExecutor,
		_store:           store,
		_dbName:          dbName,
		_logger:          logger,
	}

	fn := func(entity interface{}) string {
		panic("Shouldn't be generating new ids here")
	}
	c := res._requestExecutor.GetConventions()
	res._generateEntityIdOnTheClient = newGenerateEntityIDOnTheClient(c, fn)
	return res
}

func (b *SubscriptionBatch) initialize(batch []*SubscriptionConnectionServerMessage) (string, error) {
	b.Items = nil

	lastReceivedChangeVector := ""

	for _, item := range batch {
		curDoc := item.Data
		metadataI, ok := curDoc[MetadataKey]
		if !ok {
			return "", throwRequired("@metadata field")
		}

		metadata := metadataI.(map[string]interface{})
		id, ok := jsonGetAsText(metadata, MetadataID)
		if !ok {
			return "", throwRequired("@id field")
		}
		changeVector, ok := jsonGetAsText(metadata, MetadataChangeVector)
		if !ok {
			return "", throwRequired("@change-vector field")
		}
		lastReceivedChangeVector = changeVector
		if b._logger != nil {
			b._logger.Printf("Got %s (change vector: [%s], size: %d)", id, lastReceivedChangeVector, len(curDoc))
		}
		var instance interface{}

		if item.Exception == "" {
			if b._clazz == reflect.TypeOf(map[string]interface{}{}) {
				instance = curDoc
			} else {
				if b._revisions {
					// parse outer object manually as Previous/Current has PascalCase
					previous := curDoc["Previous"]
					current := curDoc["Current"]
					revision := &Revision{}
					//c := b._requestExecutor.GetConventions()
					if current != nil {
						doc := current.(map[string]interface{})
						v, err := entityToJSONConvertToEntity(b._clazz, id, doc)
						if err != nil {
							return "", err
						}
						revision.Current = v
					}
					if previous != nil {
						doc := previous.(map[string]interface{})
						v, err := entityToJSONConvertToEntity(b._clazz, id, doc)
						if err != nil {
							return "", err
						}
						revision.Previous = v
					}
					instance = revision
				} else {
					var err error
					instance, err = entityToJSONConvertToEntity(b._clazz, id, curDoc)
					if err != nil {
						return "", err
					}
				}
			}

			if stringIsNotEmpty(id) {
				b._generateEntityIdOnTheClient.trySetIdentity(instance, id)
			}
		}
		itemToAdd := &SubscriptionBatchItem{
			changeVector:     changeVector,
			ID:               id,
			rawResult:        curDoc,
			rawMetadata:      metadata,
			_result:          instance,
			exceptionMessage: item.Exception,
		}
		b.Items = append(b.Items, itemToAdd)
	}
	return lastReceivedChangeVector, nil
}

func throwRequired(name string) error {
	return newIllegalStateError("Document must have a " + name)
}
