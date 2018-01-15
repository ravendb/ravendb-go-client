package store

import (
	"github.com/ravendb-go-client/data"
	"io"
	"errors"
	"github.com/ravendb-go-client/http"
	"time"
	"github.com/ravendb-go-client/http/commands"
	"github.com/ravendb-go-client/tools"
	"github.com/ravendb-go-client/hilo"
	"github.com/ravendb-go-client/tools/types"
)
type DocumentStore struct {
	Conventions     *data.DocumentConvention
	urls            []string
	DefaultDBName   string
	certificate     io.ByteReader
	requestExecutor *http.RequestExecutor
	generator       *hilo.MultiDatabaseHiLoKeyGenerator
	operations      *OperationExecutor
	initialize      bool
	admin           *AdminOperationExecutor
}
// rename parameter database to DefaultDBName
//  add method initialize theare
func NewDocumentStore(urls []string, dbName string, certificate io.ByteReader) (*DocumentStore, error) {

	if (urls == nil) || len(urls) == 0 {
		return nil, errors.New("Document store URLs cannot be empty")
	}
	if dbName == "" {
		return nil, errors.New("DefaultDBName is not valid")
	}

	ref := &DocumentStore{}
	ref.urls = urls
	ref.DefaultDBName = dbName
	ref.Conventions = data.NewDocumentConvention()
	ref.certificate = certificate
	//TODO: implamate later
	//ref.lock = Lock()
//ref.operations, err = NewOperationExecutor(ref, ref.DefaultDBName)
	ref.generator = hilo.NewMultiDatabaseHiLoKeyGenerator(ref.DefaultDBName, ref.urls[0], ref.Conventions)
	ref.admin = NewAdminOperationExecutor(ref, dbName)
	//ref.subscription = DocumentSubscriptions(ref)

	return ref, nil
}

func (ref *DocumentStore) GetCertificate() io.ByteReader {
	return ref.certificate
}

func (ref *DocumentStore) GetOperations() *OperationExecutor{
	ref.assertInitialize()
	return ref.operations
}
//func (ref *DocumentStore) __enter__() {
//	return ref
//}
//func (ref *DocumentStore) __exit__(exc_type, exc_val, exc_tb) {
//	if ref.generator != nil {
//		ref.generator.return_unused_range()
//	}
//}
func (ref *DocumentStore) GetRequestExecutor(db_name string) (*http.RequestExecutor, error) {
	if db_name == "" {
		db_name = ref.DefaultDBName
	}

	//with ref.lock:
	//if db_name not in ref.requestExecutor:
	//ref.requestExecutor[db_name] = RequestsExecutor.create(ref.urls, db_name, ref.certificate,
	//ref.Conventions)
	//return ref.requestExecutor[db_name]

	return http.CreateForSingleNode(ref.urls[0], db_name)
}
func (ref *DocumentStore) assertInitialize() error {
	if !ref.initialize {
		return errors.New("You cannot open a session || access the DefaultDBName commands before initializing the document store.Did you forget calling initialize()?")
	}
	return nil
}
func (ref *DocumentStore) OpenSession(database string, requests_executor *http.RequestExecutor) (*DocumentSession, error){
	ref.assertInitialize()
	sessionId := tools.Uuid4()
	if requests_executor == nil {
		requests_executor, err := ref.GetRequestExecutor(database)
		if err != nil {
			return nil, err
		}
		ref.requestExecutor = requests_executor
	} else {
		ref.requestExecutor = requests_executor
	}
	return NewDocumentSession(database, ref, ref.requestExecutor, sessionId), nil
}
func (ref *DocumentStore) generate_id(dbName string, entity types.Document) string{
	if ref.generator != nil{
		return ref.generator.GenerateDocumentKey(dbName, entity)
	}
	return ""
}


//# ------------------------------Operation executors ---------------------------->
type AdminOperationExecutor struct {
	store           *DocumentStore
	dbName          string
	server          *ServerOperationExecutor
	requestExecutor *http.RequestExecutor
}
func NewAdminOperationExecutor(documentStore *DocumentStore, dbName string) *AdminOperationExecutor {
	ref := &AdminOperationExecutor{}
	ref.store = documentStore
	if dbName > "" {
		ref.dbName = dbName
	} else {
		ref.dbName = documentStore.DefaultDBName
	}
	ref.server = NewServerOperationExecutor(ref.store)

	return ref
}

func (ref *AdminOperationExecutor) GetRequestExecutor() *http.RequestExecutor {
	if ref.requestExecutor == nil {
		ref.requestExecutor,_ = ref.store.GetRequestExecutor(ref.dbName)
	}
	return ref.requestExecutor
}
func (ref *AdminOperationExecutor) send(operation commands.IRavenRequestable) (*http.Response,error){
//if operation_type := operation.GetOperation(); operation_type != "AdminOperation" {
//	return nil, errors.New("operation type cannot be " + operation_type + " need to be Operation")
//}
//command = operation.get_command(ref.requestExecutor.convention)
return ref.requestExecutor.ExecuteOnCurrentNode(operation, false)
}

type ServerOperationExecutor struct {
	store *DocumentStore
	requestExecutor *http.RequestExecutor
}
func NewServerOperationExecutor(documentStore *DocumentStore) *ServerOperationExecutor {
	ref := &ServerOperationExecutor{}
	ref.store = documentStore

	return ref
}

func (ref *ServerOperationExecutor) request_executor() (*http.RequestExecutor, error) {
	if ref.requestExecutor == nil {
		var err error
		if ref.store.Conventions.DisableTopologyUpdates {
			//                 self._request_executor = ClusterRequestExecutor.create_for_single_node(self._store.urls[0],
			//self._store.certificate)
			// todo:  add implementation like http.CreateForSingleNode(ref.store.urls[0], ref.store.certificate)

			ref.requestExecutor, err = http.CreateForSingleNode(ref.store.urls[0], ref.store.DefaultDBName)
		} else {
		//	                self._request_executor = ClusterRequestExecutor.create(self._store.urls, self._store.certificate)
		// todo: add implementation like http.Create(ref.store.urls, ref.store.certificate)
			ref.requestExecutor,err = http.Create(ref.store.urls, ref.store.DefaultDBName)
		}
		return ref.requestExecutor, err
	}

	return ref.requestExecutor, nil
}
//todo: implementation
func (ref *ServerOperationExecutor) send(operation OperationExecutor) {
//try:
//operation_type = getattr(operation, 'operation')
//if operation_type != "ServerOperation":
//return errors.New("operation type cannot be {0} need to be Operation".format(operation_type))
//except AttributeError:
//return errors.New("Invalid operation")
//
//command = operation.get_command(ref.request_executor.convention)
//return ref.request_executor.execute(command)
	}

type OperationExecutor struct {
	documentStore *DocumentStore
	database_name string
	requestExecutor *http.RequestExecutor
}
func NewOperationExecutor(documentStore *DocumentStore, database_name string) (ref *OperationExecutor, err error){
	ref = &OperationExecutor{documentStore: documentStore, database_name: database_name }
	ref.requestExecutor, err = documentStore.GetRequestExecutor(database_name)

	return
}
func (ref *OperationExecutor) wait_for_operation_complete(operation_id string, timeout time.Duration) error {
	start_time := time.Now()
	get_operation_command := commands.NewGetOperationStateCommand(operation_id)
	for {
		resp, err := ref.requestExecutor.ExecuteOnCurrentNode(get_operation_command, false)
		if err != nil || len(resp.Results) == 0 {
			return err
		}
		if (timeout > 0) && (time.Since(start_time) > timeout) {
			return errors.New("The Operation did not finish before the timeout end")
		}
		//if response["Status"] == "Completed":
		//return response
		//if response["Status"] == "Faulted":
		//return errors.New(response["Result"]["Error"])
		time.Sleep(500)
	}
}
//todo: implement
func (ref *OperationExecutor) send(operation commands.IRavenRequestable) {
//try:
//operation_type = getattr(operation, 'operation')
//if operation_type != "Operation":
//return errors.New("operation type cannot be {0} need to be Operation".format(operation_type))
//except AttributeError:
//return errors.New("Invalid operation")
//
//command = operation.get_command(ref.document_store, ref.request_executor.convention)
//return ref.request_executor.execute(command)
//
}