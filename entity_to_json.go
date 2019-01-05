package ravendb

import (
	"reflect"
)

type entityToJSON struct {
	_session           *InMemoryDocumentSessionOperations
	_missingDictionary map[interface{}]map[string]interface{}
	//private final Map<Object, Map<string, Object>> _missingDictionary = new TreeMap<>((o1, o2) -> o1 == o2 ? 0 : 1);
}

// All the listeners for this session
func newEntityToJSON(session *InMemoryDocumentSessionOperations) *entityToJSON {
	return &entityToJSON{
		_session: session,
	}
}

func (e *entityToJSON) getMissingDictionary() map[interface{}]map[string]interface{} {
	return e._missingDictionary
}

func convertEntityToJSON(entity interface{}, documentInfo *documentInfo) map[string]interface{} {
	// maybe we don't need to do anything?
	if v, ok := entity.(map[string]interface{}); ok {
		return v
	}
	jsonNode := StructToJSONMap(entity)

	entityToJSONWriteMetadata(jsonNode, documentInfo)

	tryRemoveIdentityProperty(jsonNode)

	return jsonNode
}

// TODO: verify is correct, write a test
func isTypeObjectNode(entityType reflect.Type) bool {
	var v map[string]interface{}
	typ := reflect.ValueOf(v).Type()
	return typ.String() == entityType.String()
}

// assumes v is ptr-to-struct and result is ptr-to-ptr-to-struct
func setInterfaceToValue(result interface{}, v interface{}) {
	out := reflect.ValueOf(result)
	outt := out.Type()
	outk := out.Kind()
	//fmt.Printf("outt: %s, outk: %s\n", outt, outk)
	if outk == reflect.Ptr && out.IsNil() {
		out.Set(reflect.New(outt.Elem()))
	}
	if outk == reflect.Ptr {
		out = out.Elem()
		//outt = out.Type()
		//outk = out.Kind()
	}
	//fmt.Printf("outt: %s, outk: %s\n", outt, outk)
	vin := reflect.ValueOf(v)
	//fmt.Printf("int: %s, ink: %s\n", vin.Type(), vin.Kind())
	out.Set(vin)
}

// ConvertToEntity2 converts document to a value result, matching type of result
func (e *entityToJSON) ConvertToEntity2(result interface{}, id string, document map[string]interface{}) {
	if _, ok := result.(*map[string]interface{}); ok {
		setInterfaceToValue(result, document)
		return
	}

	if _, ok := result.(map[string]interface{}); ok {
		// TODO: is this codepath ever executed?
		setInterfaceToValue(result, document)
		return
	}
	// TODO: deal with default values
	entityType := reflect.TypeOf(result)
	entity, _ := makeStructFromJSONMap(entityType, document)
	TrySetIDOnEntity(entity, id)
	setInterfaceToValue(result, entity)
}

// Converts a json object to an entity.
func (e *entityToJSON) ConvertToEntity(entityType reflect.Type, id string, document map[string]interface{}) (interface{}, error) {
	if isTypeObjectNode(entityType) {
		return document, nil
	}
	// TODO: deal with default values
	entity, err := makeStructFromJSONMap(entityType, document)
	if err != nil {
		return nil, err
	}
	TrySetIDOnEntity(entity, id)
	return entity, nil
}

func entityToJSONWriteMetadata(jsonNode map[string]interface{}, documentInfo *documentInfo) {
	if documentInfo == nil {
		return
	}

	setMetadata := false
	metadataNode := map[string]interface{}{}

	metadata := documentInfo.metadata
	metadataInstance := documentInfo.metadataInstance
	if len(metadata) > 0 {
		setMetadata = true
		for property, v := range metadata {
			v = deepCopy(v)
			metadataNode[property] = v
		}
	} else if metadataInstance != nil {
		setMetadata = true
		for key, value := range metadataInstance.EntrySet() {
			metadataNode[key] = value
		}
	}

	collection := documentInfo.collection
	if collection != "" {
		setMetadata = true

		metadataNode[MetadataCollection] = collection
	}

	if setMetadata {
		jsonNode[MetadataKey] = metadataNode
	}
}

/*
    //TBD public static object ConvertToEntity(Type entityType, string id, BlittableJsonReaderObject document, DocumentConventions conventions)

}
*/

func tryRemoveIdentityProperty(document map[string]interface{}) bool {
	delete(document, IdentityProperty)
	return true
}
