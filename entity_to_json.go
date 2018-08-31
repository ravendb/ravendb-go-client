package ravendb

import (
	"reflect"
)

type EntityToJson struct {
	_session           *InMemoryDocumentSessionOperations
	_missingDictionary map[Object]map[string]Object
	//private final Map<Object, Map<string, Object>> _missingDictionary = new TreeMap<>((o1, o2) -> o1 == o2 ? 0 : 1);
}

// All the listeners for this session
func NewEntityToJson(session *InMemoryDocumentSessionOperations) *EntityToJson {
	return &EntityToJson{
		_session: session,
	}
}

func (e *EntityToJson) getMissingDictionary() map[Object]map[string]Object {
	return e._missingDictionary
}

func EntityToJson_convertEntityToJson(entity Object, documentInfo *DocumentInfo) ObjectNode {
	// maybe we don't need to do anything?
	if v, ok := entity.(ObjectNode); ok {
		return v
	}
	jsonNode := StructToJSONMap(entity)

	EntityToJson_writeMetadata(jsonNode, documentInfo)

	tryRemoveIdentityProperty(jsonNode)

	return jsonNode
}

// TODO: verify is correct, write a test
func isTypeObjectNode(entityType reflect.Type) bool {
	var v ObjectNode
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

func (e *EntityToJson) ConvertToEntity2(result interface{}, id string, document ObjectNode) {
	entityType := reflect.TypeOf(result)
	if isTypeObjectNode(entityType) {
		setInterfaceToValue(result, document)
		return
	}
	// TODO: deal with default values
	entity, _ := MakeStructFromJSONMap(entityType, document)
	TrySetIDOnEntity(entity, id)
	setInterfaceToValue(result, entity)
}

// Converts a json object to an entity.
func (e *EntityToJson) ConvertToEntity(entityType reflect.Type, id string, document ObjectNode) Object {
	if isTypeObjectNode(entityType) {
		return document
	}
	// TODO: deal with default values
	entity, _ := MakeStructFromJSONMap(entityType, document)
	TrySetIDOnEntity(entity, id)
	return entity
}

func EntityToJson_writeMetadata(jsonNode ObjectNode, documentInfo *DocumentInfo) {
	if documentInfo == nil {
		return
	}

	setMetadata := false
	metadataNode := ObjectNode{}

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

		metadataNode[Constants_Documents_Metadata_COLLECTION] = collection
	}

	if setMetadata {
		jsonNode[Constants_Documents_Metadata_KEY] = metadataNode
	}
}

/*
    //TBD public static object ConvertToEntity(Type entityType, string id, BlittableJsonReaderObject document, DocumentConventions conventions)

}
*/

func tryRemoveIdentityProperty(document ObjectNode) bool {
	delete(document, IdentityProperty)
	return true
}
