package ravendb

import (
	"log"
	"reflect"
)

// SubscriptionBatchItem describes a single result from subscription
type SubscriptionBatchItem struct {
	Result       interface{}
	ErrorMessage string
	ID           string
	ChangeVector string

	RawResult   map[string]interface{}
	RawMetadata map[string]interface{}
	Metadata    *MetadataAsDictionary
}

func (i *SubscriptionBatchItem) throwItemProcessException() error {
	return newIllegalStateError("Failed to process document " + i.ID + " with Change Vector " + i.ChangeVector + " because: \n" + i.ErrorMessage)
}

// GetResult sets result to to the value of that batch item
func (i *SubscriptionBatchItem) GetResult(result interface{}) error {
	if i.ErrorMessage != "" {
		return i.throwItemProcessException()
	}
	if v, ok := result.(*map[string]interface{}); ok {
		*v = i.RawResult
		return nil
	}
	if err := checkValidLoadArg(result, "result"); err != nil {
		return err
	}
	return setInterfaceToValue(result, i.Result)
}

func (i *SubscriptionBatchItem) GetMetadata() *MetadataAsDictionary {
	if i.Metadata == nil {
		i.Metadata = NewMetadataAsDictionary(i.RawMetadata, nil, "")
	}

	return i.Metadata
}

// SubscriptionBatch describes a bunch of results for subscription
type SubscriptionBatch struct {
	clazz           reflect.Type
	revisions       bool
	requestExecutor *RequestExecutor
	store           *DocumentStore
	dbName          string

	logger                      *log.Logger
	generateEntityIdOnTheClient *generateEntityIDOnTheClient

	Items []*SubscriptionBatchItem
}

func (b *SubscriptionBatch) OpenSession() (*DocumentSession, error) {
	sessionOptions := &SessionOptions{
		Database:        b.dbName,
		RequestExecutor: b.requestExecutor,
	}
	return b.store.OpenSessionWithOptions(sessionOptions)
}

func newSubscriptionBatch(clazz reflect.Type, revisions bool, requestExecutor *RequestExecutor, store *DocumentStore, dbName string, logger *log.Logger) *SubscriptionBatch {
	res := &SubscriptionBatch{
		clazz:           clazz,
		revisions:       revisions,
		requestExecutor: requestExecutor,
		store:           store,
		dbName:          dbName,
		logger:          logger,
	}

	fn := func(entity interface{}) (string, error) {
		panic("Shouldn't be generating new ids here")
	}
	c := res.requestExecutor.GetConventions()
	res.generateEntityIdOnTheClient = newGenerateEntityIDOnTheClient(c, fn)
	return res
}

func (b *SubscriptionBatch) initialize(batch []*subscriptionConnectionServerMessage) (string, error) {
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
		if b.logger != nil {
			b.logger.Printf("Got %s (change vector: [%s], size: %d)", id, lastReceivedChangeVector, len(curDoc))
		}
		var instance interface{}

		if item.Exception == "" {
			if b.clazz == reflect.TypeOf(map[string]interface{}{}) {
				instance = curDoc
			} else {
				if b.revisions {
					// parse outer object manually as Previous/Current has PascalCase
					previous := curDoc["Previous"]
					current := curDoc["Current"]
					revision := &Revision{}
					//c := b._requestExecutor.GetConventions()
					if current != nil {
						doc := current.(map[string]interface{})
						v, err := entityToJSONConvertToEntity(b.clazz, id, doc)
						if err != nil {
							return "", err
						}
						revision.Current = v
					}
					if previous != nil {
						doc := previous.(map[string]interface{})
						v, err := entityToJSONConvertToEntity(b.clazz, id, doc)
						if err != nil {
							return "", err
						}
						revision.Previous = v
					}
					instance = revision
				} else {
					var err error
					instance, err = entityToJSONConvertToEntity(b.clazz, id, curDoc)
					if err != nil {
						return "", err
					}
				}
			}

			if stringIsNotEmpty(id) {
				b.generateEntityIdOnTheClient.trySetIdentity(instance, id)
			}
		}
		itemToAdd := &SubscriptionBatchItem{
			ChangeVector: changeVector,
			ID:           id,
			RawResult:    curDoc,
			RawMetadata:  metadata,
			Result:       instance,
			ErrorMessage: item.Exception,
		}
		b.Items = append(b.Items, itemToAdd)
	}
	return lastReceivedChangeVector, nil
}

func throwRequired(name string) error {
	return newIllegalStateError("Document must have a " + name)
}
