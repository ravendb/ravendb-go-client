package store

import (
	"github.com/ravendb-go-client/http"
	"github.com/ravendb-go-client/data"
	"github.com/ravendb-go-client/tools/types"
	"errors"
	"github.com/ravendb-go-client/http/commands"
	"fmt"
)

type SaveChangesData struct {
	commands []string
	deferredCommandCount int
	entities []string
}
func NewSaveChangesData(commands []string, deferredCommandCount int, entities ...string) (*SaveChangesData) {
	ref := &SaveChangesData{}
	ref.commands = commands
	ref.entities = entities 
	ref.deferredCommandCount = deferredCommandCount
	
	return ref
}
// DocumentSession - not fill inplemented
type DocumentSession struct {
	id_value, numberOfRequestsInSession             uint
	database                                        string
	documentStore                                   *DocumentStore
	requestsExecutor                                *http.RequestExecutor
	sessionId                                       uint64
	documentsById                                   types.TDocByID
	includedDocumentsById                           types.SETstr
	documentsByEntity                               types.TDocByEntities
	deletedEntities, knownMissingIds, deferCommands types.SETstr
	query                                           *Query
	conventions                                     *data.DocumentConvention
	advanced                                        *Advanced
}
//Implements Unit of Work for accessing the RavenDB server
//
//@param str DefaultDBName: the name of the DefaultDBName we open a session to
//@param DocumentStore documentStore: the store that we work on
// parameter kwargs удалил как неиспользуемый
func NewDocumentSession(dbName string, docStore *DocumentStore, reqExecutor *http.RequestExecutor, sessionId uint64) *DocumentSession{
	ref := &DocumentSession{}
	ref.sessionId = sessionId
	ref.database = dbName
	ref.documentStore = docStore
	ref.requestsExecutor = reqExecutor
	ref.advanced = NewAdvanced(ref)

	return ref
}
func (obj DocumentSession) GetConvention() *data.DocumentConvention{
	return obj.conventions
}

func (ref *DocumentSession) RequestsExecutor() *http.RequestExecutor{
	return ref.requestsExecutor
}

func (ref *DocumentSession) GetNumberOfRequestsInSession() uint{
	return ref.numberOfRequestsInSession
}

func (ref *DocumentSession) GetDocumentByEntity() types.TDocByEntities{
	return ref.documentsByEntity
}

func (ref *DocumentSession) GetDeletedEntities() types.SETstr{
	return ref.deletedEntities
}

func (ref *DocumentSession) documents_by_id() types.TDocByID{
	return ref.documentsById
}

func (ref *DocumentSession) GetKnownMissingIds() types.SETstr{
	return ref.knownMissingIds
}

func (ref *DocumentSession) GetIncludedDocumentsById() types.SETstr{
	return ref.includedDocumentsById
}

func (ref *DocumentSession) GetConventions() *data.DocumentConvention{
	return ref.documentStore.Conventions
}

func (ref *DocumentSession) GetQuery() *Query{
	if ref.query == nil {
		ref.query = NewQuery()
	}
	return ref.query
}
func (ref *DocumentSession) save_includes(includes types.Set) {
	ref.includedDocumentsById.Append(includes)
}

func (ref *DocumentSession) saveEntity(key string, entity, original_metadata, metadata, document string, force_concurrency_check bool) {
	if key > "" {
		ref.knownMissingIds.Delete(key)

		if _, ok := ref.documentsById[key]; !ok {
			ref.documentsById[key] = entity

			ref.documentsByEntity[entity] = &types.TDocByEntity{
				Original_value:    "document.copy()", Metadata: metadata,
				Original_metadata: original_metadata, Change_vector: `metadata.get("change_vector", None)`,
				Key:               key, Force_concurrency_check: force_concurrency_check,
			}
		}
	}
}
//todo: complete this method
func (ref *DocumentSession) convertAndSaveEntity(key, document, object_type, nested_object_types string) {
	if _, ok := ref.documentsById[key]; !ok {
		//entity, metadata, original_metadata = Utils.convert_to_entity(document, object_type, ref.Conventions,
		//nested_object_types)
		//ref.saveEntity(key, entity, original_metadata, metadata, document)
	}
}
// todo: nedd to full refactoring
func (ref *DocumentSession) multiLoad(keys []string,  object_type interface{}, includes []string, nested_object_types []interface{}) []string {
	if len(keys) == 0 {
		return nil
	}

	//ids_of_not_existing_object = set(keys)
	//if not includes:
	//ids_in_includes = [key for key in ids_of_not_existing_object if key in ref.included_documents_by_id]
	//if len(ids_in_includes) > 0:
	//for include in ids_in_includes:
	//ref.convertAndSaveEntity(include, ref.included_documents_by_id[include], object_type,
	//nested_object_types)
	//ref.included_documents_by_id.pop(include)
	//
	//ids_of_not_existing_object = [key for key in ids_of_not_existing_object if
	//key not in ref.documents_by_id]
	//
	//ids_of_not_existing_object = [key for key in ids_of_not_existing_object if key not in ref.known_missing_ids]
	//
	//if len(ids_of_not_existing_object) > 0:
	//ref.IncrementRequestsCount()
	//command = GetDocumentCommand(ids_of_not_existing_object, includes)
	//response = ref.requestsExecutor.execute(command)
	//if response:
	//results = response["Results"]
	//includes = response["Includes"]
	//for i in range (0, len(results)):
	//if results[i] == "" {
	//ref.known_missing_ids.add(ids_of_not_existing_object[i])
	//continue
	//ref.convertAndSaveEntity(ids_of_not_existing_object[i], results[i], object_type,
	//nested_object_types)
	//ref.save_includes(includes)
	//return [None if key in ref.known_missing_ids else ref.documents_by_id[
	//key] if key in ref.documents_by_id else None for key in keys]
	
	return nil
}
// todo: again bind to previous method and create single processing and loop where it is needed
//@param key_or_keys: Identifier of a document that will be loaded.
//:type str || list
//@param includes: The path to a reference inside the loaded documents can be list (property name)
//:type list || str
//@param object_type: The class we want to get
//:type classObj:
//@param nested_object_types: A dict of classes for nested object the key will be the name of the class && the
//value will be the object we want to get for that attribute
//:type str
//@return: instance of object_type || None if document with given Id does not exist.
//:rtype:object_type || None
func (ref *DocumentSession) load(key_or_keys []string, object_type interface{}, includes []string, nested_object_types []interface{}) (*DocumentSession, error) {
	if key_or_keys != nil {
		return nil, errors.New("None || empty key is invalid")
	}
	//if isinstance(key_or_keys, list) {
	//	return ref.multiLoad(key_or_keys, object_type, includes, nested_object_types)
	//}
	//
	//if key_or_keys in ref.known_missing_ids:
	//return None
	//if key_or_keys in ref.documents_by_id && not includes:
	//return ref.documents_by_id[key_or_keys]
	//
	//if key_or_keys in ref.included_documents_by_id:
	//ref.convertAndSaveEntity(key_or_keys, ref.included_documents_by_id[key_or_keys], object_type,
	//nested_object_types)
	//ref.included_documents_by_id.pop(key_or_keys)
	//if not includes:
	//return ref.documents_by_id[key_or_keys]
	//
	//ref.IncrementRequestsCount()
	//command = GetDocumentCommand(key_or_keys, includes = includes)
	//response = ref.requestsExecutor.execute(command)
	//if response:
	//result = response["Results"]
	//includes = response["Includes"]
	//if len(result) == 0 || result[0] == "" {
	//ref.known_missing_ids.add(key_or_keys)
	//return None
	//ref.convertAndSaveEntity(key_or_keys, result[0], object_type, nested_object_types)
	//ref.save_includes(includes)
	//return ref.documents_by_id[key_or_keys] if key_or_keys in ref.documents_by_id else None
	//}
	return nil, nil
}
func (ref *DocumentSession) deleteByEntity(entity string) error {
	if entity == "" {
		return errors.New("None entity is invalid")
	}
	if item, ok := ref.documentsByEntity[entity]; ok {
		return errors.New(entity + " is not associated with the session, cannot delete unknown entity instance")
	} else if "Raven-Read-Only" == item.Original_metadata {
		return errors.New(entity + " is marked as read only && cannot be deleted")
	} else {
		ref.includedDocumentsById.Add(item.Key)
		ref.knownMissingIds.Add(item.Key)
		ref.deletedEntities.Add(entity)
	}
	
	return nil
}
//todo: split in two separate methods
//@param key_or_entity:can be the key || the entity we like to delete
func (ref *DocumentSession) delete(key_or_entity string) error {
		if key_or_entity  == "" {
			return errors.New("None key is invalid")
		}
		//if not isinstance(key_or_entity, str):
		//ref.deleteByEntity(key_or_entity)
		//return
		if entity, ok := ref.documentsById[key_or_entity]; ok {
			if ref.hasChange(entity) {
				return errors.New("Can't delete changed entity using identifier. Use deleteByEntity(entity) instead.")
			}
			if doc, ok := ref.documentsByEntity[entity]; !ok {
				return errors.New(entity + " is not associated with the session, cannot delete unknown entity instance")
			} else if "Raven-Read-Only" == doc.Original_metadata {
				return errors.New(entity + " is marked as read only && cannot be deleted")
			}
			ref.deleteByEntity(entity)
		} else {
			ref.knownMissingIds.Add(key_or_entity)
			//ref.includedDocumentsById.Add(key_or_entity)
			//ref.deferCommands.Add(commands_data.DeleteCommandData(key_or_entity))
		}
	return nil
	}
func (ref *DocumentSession) assert_no_non_unique_instance(entity, key string) {
	//if not (key is None || key.endswith("/") || key not in ref.documents_by_id
	//or ref.documents_by_id[key] is entity):
	//return nil, errors.New exceptions.NonUniqueObjectException(
	//"Attempted to associate a different object with id '{0}'.".format(key)) 
}
//@param entity: Entity that will be stored
//:type object:
//@param key: Entity will be stored under this key, (None to generate automatically)
//:type str:
//@param change_vector: Current entity change_vector, used for concurrency checks (null to skip check)
//:type str
func (ref *DocumentSession) store(entity, key, change_vector string) error {

	if entity == "" {
		return errors.New("None entity value is invalid")
	}

	force_concurrency_check := ref.getConcurrencyCheckMode(entity, key, change_vector)
	if entity, ok := ref.documentsByEntity[entity]; ok {
		if change_vector > "" {
			entity.Change_vector = change_vector
		}
		entity.Force_concurrency_check = force_concurrency_check
		return nil
	}

	//if key == "" {
	//	entity_id = GenerateEntityIdOnTheClient.try_get_id_from_instance(entity)
	//} else {
	//	GenerateEntityIdOnTheClient.try_set_id_on_entity(entity, key)
	//	entity_id = key
	//}
	//
	//ref.assert_no_non_unique_instance(entity, entity_id)
	//
	//if entity_id >0 {
	//	entity_id = ref.documentStore.generate_id(ref.database, entity)
	//	GenerateEntityIdOnTheClient.try_set_id_on_entity(entity, entity_id)
	//}
	//for command in ref.defer_commands:
	//if command.key == entity_id {
	//	return errors.New("Can't store document, there is a deferred command registered for this document in the session. "
	//	"Document id: " + entity_id)
	//}
	//
	//if entity in ref.deleted_entities{
	//		return errors.New("Can't store object, it was already deleted in this session.  Document id: " + entity_id)
	//	}

	//metadata = ref.Conventions.build_default_metadata(entity)
	//ref.deleted_entities.discard(entity)
	//ref.saveEntity(entity_id, entity, {}, metadata, {}, force_concurrency_check=force_concurrency_check)
	return nil
}
// todo: sort out - force_concurrency_check obviously bool based on lines above but its a string here!
func (ref *DocumentSession) getConcurrencyCheckMode(entity, key, change_vector string) bool {
	
    defaultResult := "forced"
    if change_vector == "" {
		defaultResult = "disabled"
	}

	if key == "" {
	    fmt.Print(defaultResult)
		return  false
	}
	//todo: dig into upper layer
	//if change_vector == "" {
	//	if key == "" {
	//		entity_key = GenerateEntityIdOnTheClient.try_get_id_from_instance(entity)
	//		return "forced"
	//		if entity_key is
	//		None else "auto"
	//	}
	//	return "auto"
	//}
	//defaultResult
	return false
	}
func (ref *DocumentSession) saveChanges() error{
	data := NewSaveChangesData(ref.deferCommands.ToSlice(), len(ref.deferCommands))
	ref.deferCommands.Clear()
	ref.prepareForDeleteCommands(data)
	ref.prepareForPutsCommands(data)
	if len(data.commands) > 0 {

		ref.IncrementRequestsCount()
		batch_command := commands.NewBatchCommand(data.commands)
		batch_result, err := ref.requestsExecutor.ExecuteOnCurrentNode(batch_command, true)
		if err != nil {
			return err
		}else if batch_result == nil {
			return errors.New("Cannot call Save Changes after the document store was disposed.")
		}
		ref.updateBatchResult(batch_result, data)
	}
			
	return nil
}
func (ref *DocumentSession) updateBatchResult(batch_result []byte, data *SaveChangesData) {

	// todo: for implement adding JSON marsghaling
	//i := data.deferredCommandCount
	//batch_result_length = len(batch_result)
	//while i < batch_result_length:
	//item = batch_result[i]
	//if item["Type"] == "PUT":
	//entity = data.entities[i - data.deferred_command_count]
	//if entity in ref.documentsByEntity:
	//ref.documents_by_id[item["@id"]] = entity
	//document_info = ref.documentsByEntity[entity]
	//document_info["change_vector"] = ["change_vector"]
	//item.pop("Type", None)
	//document_info["original_metadata"] = item.copy()
	//document_info["metadata"] = item
	//document_info["original_value"] = entity.__dict__.copy()
	//i += 1
}
func (ref *DocumentSession) prepareForDeleteCommands(data *SaveChangesData) {
	keys_to_delete := ref.deletedEntities.ToSlice()
	{
		//	keys_to_delete = append(keys_to_delete, entity) //]["key"])
		//}

		for _, key := range keys_to_delete {
			existing_entity, change_vector := "", ""
			if existing_entity, ok := ref.documentsById[key]; ok {
				if _, ok := ref.documentsByEntity[existing_entity]; ok && ref.advanced.use_optimistic_concurrency {
					change_vector = ref.documentsByEntity[existing_entity].Change_vector
				}
				delete(ref.documentsByEntity, existing_entity)
				delete(ref.documentsById, key)
			}
			data.entities = append(data.entities, existing_entity)
			//todo: implement command_data
			data.commands = append(data.commands, "commands_data.DeleteCommandData(key, change_vector)"+change_vector)
		}
		ref.deletedEntities.Clear()
	}
}
func (ref *DocumentSession) prepareForPutsCommands(data *SaveChangesData) {
	for name, entity := range ref.documentsByEntity {
		if ref.hasChange(name) {
			key      := entity.Key
			metadata := entity.Metadata
			change_vector := ""
			if ref.advanced.use_optimistic_concurrency && entity.Force_concurrency_check {
				//|| entity.force_concurrency_check"] == "forced":
				change_vector = entity.Change_vector
			}
			data.entities = append(data.entities, name)
			if key > "" {
				delete(ref.documentsById, key)
				//document = entity.__dict__.copy()
				//document.pop('Id',None)
			}
			data.commands = append(data.commands, "commands_data.PutCommandData(key," + change_vector +", document" + metadata)
		}
	}
}
func (obj DocumentSession) hasChange(entityName string) bool {
	entity := obj.documentsByEntity[entityName]
	
	return entity.Original_metadata != entity.Metadata || entity.Original_value != "entity.__dict__"
}
func (ref *DocumentSession) IncrementRequestsCount() error{
	ref.numberOfRequestsInSession ++
	if ref.numberOfRequestsInSession > ref.conventions.MaxNumberOfRequestsPerSession {
		return errors.New( fmt.Sprintf(`The maximum number of requests (%d) allowed for this session has been reached. Raven limits the number \
	of remote calls that a session is allowed to make as an early warning system. Sessions are expected to \
	be short lived, && Raven provides facilities like batch saves (call saveChanges() only once).\
	You can increase the limit by setting DocumentConvention.\
	MaxNumberOfRequestsPerSession || MaxNumberOfRequestsPerSession, but it is advisable \
	that you'll look into reducing the number of remote calls first, \
	since that will speed up your application significantly && result in a\
	more responsive application.`, ref.conventions.MaxNumberOfRequestsPerSession))
	}
	return nil
}

type Advanced struct {
	session                    *DocumentSession
	use_optimistic_concurrency bool
}
func NewAdvanced(session *DocumentSession) *Advanced {
	ref := &Advanced{}
	ref.session = session
	return ref
}
func (ref *Advanced) numberOfRequestsInSession() uint {
	return ref.session.numberOfRequestsInSession
}
func (ref *Advanced) get_document_id(instance string) string {
	if instance > "" {

		if instance, ok := ref.session.documentsByEntity[instance]; ok {
			return instance.Key
		}
	}
	return ""
}
//The document store associated with this session
func (ref *Advanced) documentStore() *DocumentStore {
	return ref.session.documentStore
}
func (ref *Advanced) clear() {
	ref.session.documentsByEntity.Clear()
	ref.session.documentsById.Clear()
	ref.session.deletedEntities.Clear()
	ref.session.includedDocumentsById.Clear()
	ref.session.knownMissingIds.Clear()
}