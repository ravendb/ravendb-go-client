package store

import (
	"github.com/ravendb-go-client/http"
	"github.com/ravendb-go-client/data"
	"github.com/ravendb-go-client/tools/types"
	"errors"
	"github.com/ravendb-go-client/http/commands"
	"fmt"
	"reflect"
	"github.com/ravendb-go-client/tools"
)

type SaveChangesData struct {
	commands []commands.ICommandData
	deferredCommandCount int
	entities []*interface{}
}
func NewSaveChangesData(commands []commands.ICommandData, deferredCommandCount int, entities ...*interface{}) (*SaveChangesData) {
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
	deletedEntities									[]*interface{}
	knownMissingIds									types.SETstr
	deferCommands 									[]commands.ICommandData
	query                                           *Query
	conventions                                     *data.DocumentConvention
	advanced                                        *Advanced
}
//Implements Unit of Work for accessing the RavenDB server
//
//@param str DefaultDBName: the name of the DefaultDBName we open a session to
//@param DocumentStore documentStore: the store that we work on
func NewDocumentSession(dbName string, docStore *DocumentStore, reqExecutor *http.RequestExecutor, sessionId uint64) *DocumentSession{
	ref := &DocumentSession{}
	ref.sessionId = sessionId
	ref.database = dbName
	ref.documentStore = docStore
	ref.requestsExecutor = reqExecutor
	ref.advanced = NewAdvanced(ref)
	ref.conventions = data.NewDocumentConvention()
	ref.documentsByEntity = make(types.TDocByEntities)
	ref.documentsById = make(types.TDocByID)
	ref.knownMissingIds = make(types.SETstr)

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

func (ref *DocumentSession) GetDeletedEntities() []*interface{}{
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
func (ref *DocumentSession) SaveIncludes(includes types.Set) {
	ref.includedDocumentsById.Append(includes)
}

func (ref *DocumentSession) saveEntity(key string, entity interface{}, original_metadata , metadata data.Metadata, document interface{}, force_concurrency_check bool) {
	if key > "" {
		ref.knownMissingIds.Delete(key)

		if _, ok := ref.documentsById[key]; !ok {
			docRef := &types.Document{
				OriginalValue:    document, Metadata: &metadata,
				OriginalMetadata: &original_metadata, ChangeVector: []string{`metadata.get("change_vector", None)`},
				Key:              key, ForceConcurrencyCheck: force_concurrency_check,
			}
			ref.documentsById[key] = docRef
			ref.documentsByEntity[&entity] = docRef
		}
	}
}

func (ref *DocumentSession) convertAndSaveEntity(key string, document interface{}) {
	if _, ok := ref.documentsById[key]; !ok {
		entity, metadata, originalMetadata, _ := tools.ConvertToEntity(document)

		ref.saveEntity(key, entity, originalMetadata, metadata, document, false)
	}
}
func (ref *DocumentSession) multiload(keys []string,  object_type interface{}, includes []string, nested_object_types []interface{}) ([]*types.Document, error) {

	documents := make([]*types.Document, len(keys))

	if keys == nil || len(keys) == 0 {
		return nil, errors.New("None or empty key is invalid")
	}

	idsOfNotExistingObject := types.NewSETstrFromArray(keys)

	if includes == nil || len(includes) == 0{
		idsInIncludes := make(types.SETstr, 0)
		for key, _ := range *idsOfNotExistingObject{
			if ref.includedDocumentsById.HasKey(key){
				idsInIncludes.Add(key)
			}
		}
		if len(idsInIncludes) > 0{
			for include , _ := range idsInIncludes{
				ref.convertAndSaveEntity(include, ref.includedDocumentsById[include])
				ref.includedDocumentsById.Delete(include)
			}
		}
		for key, _ := range *idsOfNotExistingObject{
			if _, ok := ref.documentsById[key]; !ok{
				idsOfNotExistingObject.Add(key)
			}
		}
	}
	for key, _ := range *idsOfNotExistingObject{
		if !ref.knownMissingIds.HasKey(key){
			idsOfNotExistingObject.Add(key)
		}
	}

	if len(*idsOfNotExistingObject) > 0{
		ref.IncrementRequestsCount()
		command, _ := commands.NewGetDocumentCommand(idsOfNotExistingObject.ToSlice(), includes, false)
		response, err := ref.requestsExecutor.ExecuteOnCurrentNode(command, true)
		if err != nil{
			return nil, err
		}
		for i, v := range response.Results{
			if v == nil{
				ref.knownMissingIds.Add(idsOfNotExistingObject.ToSlice()[i])
			}else{
				ref.convertAndSaveEntity(idsOfNotExistingObject.ToSlice()[i], v)
			}
		}
		includesSet := types.NewSETFromArray(includes)
		ref.SaveIncludes(*includesSet)
	}

	for i, key := range keys{
		if _, ok := ref.knownMissingIds[key]; ok{
			documents[i] = nil
		}else{
			documents[i] = ref.documentsById[key]
		}
	}

	return documents, nil
}
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
func (ref *DocumentSession) Load(keys []string, object_type interface{}, includes []string, nested_object_types []interface{}) ([]*types.Document, error) {
	if keys == nil || len(keys) == 0{
		return nil, errors.New("none or empty key is invalid")
	}
	return ref.multiload(keys, object_type, includes, nested_object_types)
}
func (ref *DocumentSession) deleteByEntity(entity *interface{}) error {
	if entity == nil {
		return errors.New("none entity is invalid")
	}
	if item, ok := ref.documentsByEntity[entity]; ok {
		return errors.New(fmt.Sprintf("%v is not associated with the session, cannot delete unknown entity instance", entity))
	} else if item.OriginalMetadata.RavenReadOnly {
		return errors.New(fmt.Sprintf("%v is marked as read only && cannot be deleted", entity))
	} else {
		ref.includedDocumentsById.Add(item.Key)
		ref.knownMissingIds.Add(item.Key)
		ref.deletedEntities = append(ref.deletedEntities, entity)
	}
	
	return nil
}
//@param key_or_entity:can be the key || the entity we like to delete
func (ref *DocumentSession) Delete(key string, entity *interface{}, expected_change_vector []string) error {
		if key  == "" && tools.IsZeroOfUnderlyingType(entity){
			return errors.New("neither key nor entity were supplied")
		}
		if !tools.IsZeroOfUnderlyingType(entity) {
			doc, ok := ref.documentsByEntity[entity]
			if ok{
				key = doc.Key
			}else{
				return errors.New("key and entity are invalid")
			}
		}
		if document, ok := ref.documentsById[key]; ok {
			if ref.hasChange(entity) {
				return errors.New("can't delete changed entity using identifier. Use deleteByEntity(entity) instead")
			}
			if entity, ok := ref.documentsByEntity.GetKeyByValue(document); !ok {
				return errors.New(fmt.Sprintf("%v is not associated with the session, cannot delete unknown entity instance", entity))
			} else if  document.OriginalMetadata.RavenReadOnly {
				return errors.New(fmt.Sprintf("%v is marked as read only && cannot be deleted", entity))
			}
			ref.deleteByEntity(entity)
		} else {
			ref.knownMissingIds.Add(key)
			ref.includedDocumentsById.Delete(key)
			commandData, _ := commands.NewDeleteCommandData(key, expected_change_vector)
			ref.deferCommands = append(ref.deferCommands, commandData)
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
func (ref *DocumentSession) store(entity *interface{}, key string, change_vector []string) error {

	if entity == nil {
		return errors.New("None entity value is invalid")
	}

	force_concurrency_check := ref.getConcurrencyCheckMode(entity, key, change_vector)
	if entity, ok := ref.documentsByEntity[entity]; ok {
		if len(change_vector) > 0 {
			entity.ChangeVector = change_vector
		}
		entity.ForceConcurrencyCheck = force_concurrency_check
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
func (ref *DocumentSession) getConcurrencyCheckMode(entity *interface{}, key string, change_vector []string) bool {
	
    defaultResult := "forced"
    if len(change_vector) == 0 {
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
func (ref *DocumentSession) SaveChanges() error{
	data := NewSaveChangesData(ref.deferCommands, len(ref.deferCommands))
	ref.deferCommands = []commands.ICommandData{}
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
func (ref *DocumentSession) updateBatchResult(batch_response *http.Response, changesData *SaveChangesData) {

	fmt.Println(batch_response.Results)

	i := changesData.deferredCommandCount
	results := batch_response.Results
	batchResultLength := len(results)
	for i < batchResultLength{
		item, ok := results[i].(map[string]interface{})
		if ok{
			if item["Type"] == "PUT"{
				entity := changesData.entities[i - changesData.deferredCommandCount]
				if doc, ok := ref.documentsByEntity[entity]; ok{
					origMetadata := item["@metadata"].(data.Metadata)
					ref.documentsById[origMetadata.Id] = doc
					documentInfo := ref.documentsByEntity[entity]
					documentInfo.ChangeVector = []string{"change_vector"}
					documentInfo.OriginalMetadata = &origMetadata
					metadata := item["@metadata"].(data.Metadata)
					documentInfo.Metadata = &metadata
					documentInfo.OriginalValue = entity
				}
			}
			i+=1
		}
	}


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
	var keysToDelete []string
	for _, entity := range ref.deletedEntities{
		keysToDelete = append(keysToDelete, ref.documentsByEntity[entity].Key)
	}
	for _, key := range keysToDelete {
		var changeVector []string
		existingDocument, ok := ref.documentsById[key]
		if ok {
			entity, ok := ref.documentsByEntity.GetKeyByValue(existingDocument)
			if ok && ref.advanced.use_optimistic_concurrency {
				changeVector = existingDocument.ChangeVector
			}
			delete(ref.documentsByEntity, entity)
			delete(ref.documentsById, key)
			data.entities = append(data.entities, entity)
		}
		commandData, _ := commands.NewDeleteCommandData(key, changeVector)
		data.commands = append(data.commands, commandData)
	}
	ref.deletedEntities = []*interface{}{}
}
func (ref *DocumentSession) prepareForPutsCommands(data *SaveChangesData) {
	for entity, document := range ref.documentsByEntity {
		if ref.hasChange(entity) {
			key      := document.Key
			metadata := document.Metadata
			var changeVector []string
			if ref.advanced.use_optimistic_concurrency && document.ForceConcurrencyCheck {
				//|| entity.force_concurrency_check"] == "forced":
				changeVector = document.ChangeVector
			}
			data.entities = append(data.entities, entity)
			if key > "" {
				delete(ref.documentsById, key)
				//document = entity.__dict__.copy()
				//document.pop('Id',None)
			}
			commandData, _ := commands.NewPutCommandData(key, *metadata, changeVector, document)
			data.commands = append(data.commands, commandData)
		}
	}
}
func (obj DocumentSession) hasChange(entity interface{}) bool {
	doc := obj.documentsByEntity[&entity]
	
	return !reflect.DeepEqual(doc.OriginalMetadata, doc.Metadata) || doc.OriginalValue != entity
}
func (ref *DocumentSession) IncrementRequestsCount() error{
	ref.numberOfRequestsInSession++
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
func (ref *Advanced) getDocumentId(instance interface{}) string {
	if &instance != nil {

		if instance, ok := ref.session.documentsByEntity[&instance]; ok {
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
	ref.session.deletedEntities = []*interface{}{}
	ref.session.includedDocumentsById.Clear()
	ref.session.knownMissingIds.Clear()
}