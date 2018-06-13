package ravendb

import "reflect"

type EntityToJson struct {
	_session           *InMemoryDocumentSessionOperations
	_missingDictionary map[Object]map[string]Object
	//private final Map<Object, Map<String, Object>> _missingDictionary = new TreeMap<>((o1, o2) -> o1 == o2 ? 0 : 1);
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
	jsonNode := structToJSONMap(entity)

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

// Converts a json object to an entity.
func (e *EntityToJson) convertToEntity(entityType reflect.Type, id String, document ObjectNode) Object {
	if isTypeObjectNode(entityType) {
		return document
	}
	// TODO: deal with default values
	entity, _ := makeStructFromJSONMap(entityType, document)
	trySetIDOnEntity(entity, id)
	return entity
	/*
		try {
			Object defaultValue = InMemoryDocumentSessionOperations.getDefaultValue(entityType);
			Object entity = defaultValue;

			String documentType =_session.getConventions().getJavaClass(id, document);
			if (documentType != null) {
				Class type = Class.forName(documentType);
				if (entityType.isAssignableFrom(type)) {
					entity = _session.getConventions().getEntityMapper().treeToValue(document, type);
				}
			}

			if (entity == defaultValue) {
				entity = _session.getConventions().getEntityMapper().treeToValue(document, entityType);
			}

			if (id != null) {
				_session.getGenerateEntityIdOnTheClient().trySetIdentity(entity, id);
			}

			return entity;
		} catch (Exception e) {
			throw new IllegalStateException("Could not convert document " + id + " to entity of type " + entityType.getName(), e);
		}
	*/
}

func EntityToJson_writeMetadata(jsonNode ObjectNode, documentInfo *DocumentInfo) {
	if documentInfo == nil {
		return
	}

	setMetadata := false
	metadataNode := ObjectNode{}

	metadata := documentInfo.getMetadata()
	metadataInstance := documentInfo.getMetadataInstance()
	if len(metadata) > 0 {
		setMetadata = true
		for property, v := range metadata {
			v = deepCopy(v)
			metadataNode[property] = v
		}
	} else if metadataInstance != nil {
		setMetadata = true
		for key, value := range metadataInstance.entrySet() {
			metadataNode[key] = value
		}
	}

	collection := documentInfo.getCollection()
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
