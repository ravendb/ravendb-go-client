package ravendb

import (
	"fmt"
	"reflect"
)

// TODO: cleanup, possibly rethink entityToJSON

type entityToJSON struct {
	session           *InMemoryDocumentSessionOperations
	missingDictionary map[interface{}]map[string]interface{}
	//private final Map<Object, Map<string, Object>> _missingDictionary = new TreeMap<>((o1, o2) -> o1 == o2 ? 0 : 1);
}

// All the listeners for this session
func newEntityToJSON(session *InMemoryDocumentSessionOperations) *entityToJSON {
	return &entityToJSON{
		session: session,
	}
}

func (e *entityToJSON) getMissingDictionary() map[interface{}]map[string]interface{} {
	return e.missingDictionary
}

func convertEntityToJSON(entity interface{}, documentInfo *documentInfo) map[string]interface{} {
	// maybe we don't need to do anything?
	if v, ok := entity.(map[string]interface{}); ok {
		return v
	}
	jsonNode := structToJSONMap(entity)

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
// TODO: ensure it doesn't panic
func setInterfaceToValue(result interface{}, v interface{}) error {
	out := reflect.ValueOf(result)
	outt := out.Type()
	//fmt.Printf("outt: %s, outk: %s\n", outt, outk)
	if outt.Kind() == reflect.Ptr && out.IsNil() {
		out.Set(reflect.New(outt.Elem()))
	}
	if outt.Kind() == reflect.Ptr {
		out = out.Elem()
		//outt = out.Type()
		//outk = out.Kind()
	}
	//fmt.Printf("outt: %s, outk: %s\n", outt, outk)
	vin := reflect.ValueOf(v)
	//fmt.Printf("int: %s, ink: %s\n", vin.Type(), vin.Kind())
	if !out.CanSet() {
		return fmt.Errorf("cannot set out %s\n", out.String())
	}
	// TODO: it can still panic. Need to lift implementaion of
	// func directlyAssignable(T, V *rtype) bool {
	// from reflect package
	out.Set(vin)
	return nil
}

// makes a copy of a map and returns a pointer to it
func mapDup(m map[string]interface{}) *map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		res[k] = v
	}
	return &res
}

// ConvertToEntity2 converts document to a value result, matching type of result
func (e *entityToJSON) convertToEntity2(result interface{}, id string, document map[string]interface{}) error {
	if _, ok := result.(**map[string]interface{}); ok {
		return setInterfaceToValue(result, mapDup(document))
	}

	if _, ok := result.(map[string]interface{}); ok {
		// TODO: is this code path ever executed?
		return setInterfaceToValue(result, document)
	}
	// TODO: deal with default values
	entityType := reflect.TypeOf(result)
	entity, err := makeStructFromJSONMap(entityType, document)
	if err != nil {
		// fmt.Printf("makeStructFromJSONMap() failed with %s\n. Wanted type: %s, document: %v\n", err, entityType, document)
		return err
	}
	trySetIDOnEntity(entity, id)
	//fmt.Printf("result is: %T, entity is: %T\n", result, entity)
	if entity == nil {
		return newIllegalStateError("decoded entity is nil")
	}
	return setInterfaceToValue(result, entity)
}

// Converts a json object to an entity.
// TODO: remove in favor of entityToJSONConvertToEntity
func (e *entityToJSON) convertToEntity(entityType reflect.Type, id string, document map[string]interface{}) (interface{}, error) {
	if isTypeObjectNode(entityType) {
		return document, nil
	}
	// TODO: deal with default values
	entity, err := makeStructFromJSONMap(entityType, document)
	if err != nil {
		return nil, err
	}
	trySetIDOnEntity(entity, id)
	return entity, nil
}

func entityToJSONConvertToEntity(entityType reflect.Type, id string, document map[string]interface{}) (interface{}, error) {
	if isTypeObjectNode(entityType) {
		return document, nil
	}
	// TODO: deal with default values
	entity, err := makeStructFromJSONMap(entityType, document)
	if err != nil {
		return nil, err
	}
	trySetIDOnEntity(entity, id)
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

/*
   public static Object convertToEntity(Class<?> entityClass, String id, ObjectNode document, DocumentConventions conventions) {
       try {
           Object defaultValue = InMemoryDocumentSessionOperations.getDefaultValue(entityClass);

           Object entity = defaultValue;

           String documentType = conventions.getJavaClass(id, document);
           if (documentType != null) {
               Class<?> clazz = Class.forName(documentType);
               if (clazz != null && entityClass.isAssignableFrom(clazz)) {
                   entity = conventions.getEntityMapper().treeToValue(document, clazz);
               }
           }

           if (entity == null) {
               entity = conventions.getEntityMapper().treeToValue(document, entityClass);
           }

           return entity;
       } catch (Exception e) {
           throw new IllegalStateException("Could not convert document " + id + " to entity of type " + entityClass);
       }
   }
*/
