package ravendb

import "reflect"

type CompareExchangeSessionValue struct {
	key                  string
	index                int64
	value                *CompareExchangeValue
	changeVector         *string
	originalValue        *CompareExchangeValue
	metadataAsDictionary *MetadataAsDictionary
	state                CompareExchangeValueState
}

func NewCompareExchangeSessionValueWithValue(value *CompareExchangeValue) (*CompareExchangeSessionValue, error) {
	var status CompareExchangeValueState
	if value.GetIndex() >= 0 {
		status = compareExchangeValueStateNone
	} else {
		status = exchangeValueStateMissing
	}

	cesv, err := NewCompareExchangeSessionValue(value.Key, value.Index, status)
	if err != nil {
		return nil, err
	}

	if value.GetIndex() > 0 {

		cesv.originalValue = NewCompareExchangeValue(value.GetKey(), value.GetIndex(), compareExchangeValueOrPrimitiveToJson(value.Value))
	}

	return cesv, nil
}

func NewCompareExchangeSessionValue(key string, index int64, state CompareExchangeValueState) (*CompareExchangeSessionValue, error) {
	if len(key) == 0 {
		return nil, newIllegalStateError("Key cannot be null")
	}

	return &CompareExchangeSessionValue{
		key:   key,
		index: index,
		state: state,
	}, nil
}

func (cesv *CompareExchangeSessionValue) Create(item interface{}) (interface{}, error) {
	cesv.assertState()

	if cesv.value != nil {
		return nil, newIllegalStateError("The compare exchange value with key '" + cesv.key + "' is already tracked.")
	}

	cesv.index = 0
	cesv.value = NewCompareExchangeValue(cesv.key, cesv.index, item)
	cesv.state = compareExchangeValueStateCreated

	return cesv.value, nil
}

func (cesv *CompareExchangeSessionValue) Delete(index int64) {
	cesv.assertState()

	cesv.index = index
	cesv.state = compareExchangeValueStateDeleted
}

func (cesv *CompareExchangeSessionValue) UpdateState(index int64) error {
	cesv.index = index
	cesv.state = compareExchangeValueStateNone

	if cesv.originalValue != nil {
		cesv.originalValue.Index = index
	}

	if cesv.value != nil {
		cesv.value.SetIndex(index)
	}

	return nil
}

func (cesv *CompareExchangeSessionValue) UpdateValue(value *CompareExchangeValue) error {
	cesv.index = value.GetIndex()
	if cesv.index >= 0 {
		cesv.state = compareExchangeValueStateNone
	} else {
		cesv.state = exchangeValueStateMissing
	}

	cesv.originalValue = value

	if cesv.value != nil {
		cesv.value.SetIndex(cesv.index)

		if cesv.value.GetValue() != nil {
			doc, ok := value.GetValue().(map[string]interface{})
			if ok == false {
				return newIllegalStateError("Cannot populate")
			}
			entity, ok := cesv.value.GetValue().(map[string]interface{})
			if ok == false {
				return newIllegalStateError("Cannot populate")
			}

			for key, val := range doc {
				entity[key] = val
			}
		}
	}

	return nil
}

func (cesv *CompareExchangeSessionValue) assertState() error {
	switch cesv.state {
	case compareExchangeValueStateNone:
		return nil
	case exchangeValueStateMissing:
		return nil
	case compareExchangeValueStateCreated:
		return newIllegalStateError("The compare exchange value with key '" + cesv.key + "' was already stored.")
	case compareExchangeValueStateDeleted:
		return newIllegalStateError("The compare exchange value with key '" + cesv.key + "' was already deleted.")
	}

	return nil
}

func (cesv *CompareExchangeSessionValue) GetValue(clazz reflect.Type, session *InMemoryDocumentSessionOperations) (*CompareExchangeValue, error) {
	switch cesv.state {
	case compareExchangeValueStateNone, compareExchangeValueStateCreated:
		if cesv.value != nil {
			return cesv.value, nil
		}

		if cesv.value != nil {
			return nil, newIllegalStateError("Value cannot be null")
		}

		var entityObject interface{}
		if cesv.originalValue != nil && cesv.originalValue.GetValue() != nil {
			entityObject = cesv.originalValue.GetValue()
		}

		ncev := NewCompareExchangeValue(cesv.key, cesv.index, entityObject)
		cesv.value = ncev
		return ncev, nil
		break
	case exchangeValueStateMissing, compareExchangeValueStateDeleted:
		return nil, nil
	}

	return nil, newIllegalStateError("'compareExchangeValueState' is unknown")
}

func (cesv *CompareExchangeSessionValue) GetCommand(session *InMemoryDocumentSessionOperations) (ICommandData, error) {
	if cesv.state == compareExchangeValueStateNone || cesv.state == compareExchangeValueStateCreated {
		if cesv.value == nil {
			return nil, nil
		}
		var err error

		entity := cesv.compareExchangeValueToJson(cesv.value.Value, session)
		entityJson, _ := entity.(map[string]interface{})

		var metadata map[string]interface{}
		if cesv.originalValue != nil && cesv.originalValue.Metadata != nil {
			metadata = cesv.originalValue.Metadata.metadata
		}

		metadataHasChanged := false

		if cesv.value.HasMetadata() && cesv.value.Metadata.IsEmpty() == false {
			if metadata == nil {
				metadataHasChanged = true
				metadata, err = cesv.prepareMetadataForPut(cesv.key, cesv.value.GetMetadata(), session.Conventions)
				if err != nil {
					return nil, err
				}
			} else {
				cesv.validateMetadataForPut(cesv.key, cesv.value.GetMetadata())

				metadataHasChanged = session.UpdateMetadataModificationsTemp(cesv.value.GetMetadata(), metadata)
			}
		}

		var entityToInsert map[string]interface{}
		if entityJson == nil || metadataHasChanged {
			object := compareExchangeValueOrPrimitiveToJson(entity)
			entityToInsert = cesv.convertEntity(metadata, object)
			entityJson = entityToInsert
		}

		newValue := NewCompareExchangeValue(cesv.key, cesv.index, entityJson)
		hasChanged := cesv.originalValue == nil || metadataHasChanged
		if hasChanged == false {
			result, err := newValue.hasChanged(cesv.originalValue)
			if err != nil {
				return nil, err
			}
			hasChanged = result
		}

		cesv.originalValue = newValue
		if hasChanged == false {
			return nil, nil
		}

		if entityToInsert == nil {
			entityToInsert = convertEntityToJSONRaw(entity, nil, false)
			entityToInsert = cesv.convertEntity(metadata, entityToInsert)
		}

		return newPutCompareExchangeCommandData(cesv.key, entityToInsert, cesv.index), nil
	}

	switch cesv.state {
	case compareExchangeValueStateDeleted:

		return newDeleteCompareExchangeCommandData(cesv.key, cesv.index), nil
	case exchangeValueStateMissing:

		return nil, nil
	}

	return nil, newIllegalStateError("Unknown state in GetCommand")
}

func (cesv *CompareExchangeSessionValue) convertEntity(metadata map[string]interface{}, value interface{}) map[string]interface{} {
	objectNode := make(map[string]interface{})
	objectNode[compareExchangeObjectFieldName] = value

	if cesv.metadataAsDictionary != nil {
		objectNode[MetadataKey] = cesv.metadataAsDictionary.metadata
	}
	return objectNode
}

type CompareExchangeValueState = int

func compareExchangeValueOrPrimitiveToJson(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	if isPrimitiveOrWrapper(reflect.TypeOf(value)) || reflect.TypeOf(value).Kind() == reflect.String || isInstanceOfArrayOfInterface(value) {
		return value
	}

	return convertEntityToJSONRaw(value, nil, true)
}

func (cesv *CompareExchangeSessionValue) compareExchangeValueToJson(value interface{}, session *InMemoryDocumentSessionOperations) interface{} {
	if value == nil {
		return nil
	}

	if isPrimitiveOrWrapper(reflect.TypeOf(value)) || reflect.TypeOf(value).Kind() == reflect.String || isInstanceOfArrayOfInterface(value) {
		return value
	}

	return convertEntityToJSONRaw(value, nil, false)
}

func (cesv *CompareExchangeSessionValue) prepareMetadataForPut(key string, metadataDictionary *MetadataAsDictionary, conventions *DocumentConventions) (map[string]interface{}, error) {
	error := cesv.validateMetadataForPut(key, metadataDictionary)
	if error != nil {
		return nil, error
	}

	return metadataDictionary.metadata, nil
}

func (cesv *CompareExchangeSessionValue) validateMetadataForPut(key string, dictionary *MetadataAsDictionary) error {
	if dictionary.ContainsKey(MetadataExpires) {
		obj, exists := dictionary.Get(MetadataExpires)
		if exists == false {
			return newIllegalStateError("The values of " + MetadataExpires + " metadata for compare exchange '" + key + " is null.")
		}

		_, isTime := obj.(Time)
		if isTime == false && reflect.TypeOf(obj).Kind() != reflect.String {
			return newIllegalStateError("The values of " + MetadataExpires + " metadata for compare exchange '" + key + " is not valid. Use the following type: Date or string.")
		}
	}

	return nil
}

type PutCompareExchangeCommandData struct {
	*CommandData
	index    int64
	document interface{}
}

func (pcecd *PutCompareExchangeCommandData) getId() string            { return pcecd.ID }
func (pcecd *PutCompareExchangeCommandData) getName() string          { return pcecd.Name }
func (pcecd *PutCompareExchangeCommandData) getChangeVector() *string { return pcecd.ChangeVector }
func (pcecd *PutCompareExchangeCommandData) getType() CommandType     { return CompareExchangePut }
func (pcecd *PutCompareExchangeCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := pcecd.baseJSON()
	res["Index"] = pcecd.index
	res["Type"] = "CompareExchangePUT"
	res["Document"] = pcecd.document
	return res, nil
}

func newPutCompareExchangeCommandData(key string, value interface{}, index int64) *PutCompareExchangeCommandData {
	return &PutCompareExchangeCommandData{
		CommandData: &CommandData{ID: key, Type: "CompareExchangePUT"},
		index:       index,
		document:    value,
	}
}

type DeleteCompareExchangeCommandData struct {
	*CommandData
	Index int64
}

func (pcecd *DeleteCompareExchangeCommandData) getId() string            { return pcecd.ID }
func (pcecd *DeleteCompareExchangeCommandData) getName() string          { return pcecd.Name }
func (pcecd *DeleteCompareExchangeCommandData) getChangeVector() *string { return pcecd.ChangeVector }
func (pcecd *DeleteCompareExchangeCommandData) getType() CommandType     { return CompareExchangeDelete }
func (pcecd *DeleteCompareExchangeCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := pcecd.baseJSON()
	res["Index"] = pcecd.Index
	res["Type"] = "CompareExchangeDELETE"
	return res, nil
}

func newDeleteCompareExchangeCommandData(key string, index int64) *DeleteCompareExchangeCommandData {
	return &DeleteCompareExchangeCommandData{
		CommandData: &CommandData{ID: key},
		Index:       index,
	}
}

const (
	compareExchangeValueStateNone    = 0
	compareExchangeValueStateCreated = 1
	compareExchangeValueStateDeleted = 1 << 1
	exchangeValueStateMissing        = 1 << 2
	compareExchangeObjectFieldName   = "Object"
)
