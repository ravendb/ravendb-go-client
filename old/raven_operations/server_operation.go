package raven_operations

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	SecurityClearance "github.com/ravendb/ravendb-go-client/data"
	"github.com/ravendb/ravendb-go-client/http/commands"
	SrvNodes "github.com/ravendb/ravendb-go-client/http/server_nodes"
	Utils "github.com/ravendb/ravendb-go-client/tools"
)

// ServerOperation abstract class  - root operations classes
type ServerOperation struct {
	operation string
	commands.RavenCommand
}

func (ref *ServerOperation) init() {
	ref.operation = "ServerOperation"
}
func (obj ServerOperation) GetOperation() string {
	return obj.operation
}

// GetCommand - base implementation can be modified by predecessors
func (ref *ServerOperation) GetCommand(conventions string) (*ServerOperation, error) {
	return ref, nil
}
func (ref *ServerOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	return nil, nil
}

//# --------------------CreateDatabaseOperation------------------------------

type tConflictSolverConfig struct {
	DatabaseResolverId  string
	ResolveByCollection string
	ResolveToLatest     bool
}
type tDatabaseRecord struct {
	DatabaseName           string
	Disabled               bool
	Encrypted              bool
	DeletionInProgress     string
	DataDirectory          string
	Topology               string
	ConflictSolverConfig   tConflictSolverConfig
	Indexes                string
	AutoIndexes            string
	Transformers           string
	Identities             string
	Settings               interface{}
	ExternalReplication    []string
	RavenConnectionStrings interface{}
	SqlConnectionStrings   interface{}
	RavenEtls              string
	SqlEtls                string
	SecuredSettings        interface{}
	Revisions              string
	Expiration             string
	PeriodicBackups        string
	Client                 string
	CustomFunctions        interface{}
}
type CreateDatabaseOperation struct {
	ServerOperation
	replicationFactor int
	databaseRecord    *tDatabaseRecord
	conventions       string
}

//	NewCreateDatabaseOperation validate parameters, create new instance & return ref
func NewCreateDatabaseOperation(dbName string, rf int, settings interface{}, scSettings interface{}) (ref *CreateDatabaseOperation, err error) {

	err = Utils.DatabaseNameValidation(dbName)
	if err != nil {
		return nil, err
	}
	ref = &CreateDatabaseOperation{}
	ref.init()
	ref.Method = "PUT"
	if rf < 1 {
		ref.replicationFactor = 1
	} else {
		ref.replicationFactor = rf
	}
	ref.databaseRecord = &tDatabaseRecord{DatabaseName: dbName}
	if settings != nil {
		ref.databaseRecord.Settings = settings
	}
	if scSettings != nil {
		ref.databaseRecord.SecuredSettings = scSettings
	}
	ref.Data = ref.databaseRecord

	return
}

//	GetCommand return ref on command struct
func (ref *CreateDatabaseOperation) GetCommand(conventions string) (*CreateDatabaseOperation, error) {
	if conventions == "" {
		return nil, errors.New("Invalid convention")
	}
	ref.conventions = conventions
	return ref, nil
}

func (ref *CreateDatabaseOperation) CreateRequest(sn SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/admin/databases?name=%s&replication-factor=%d", sn.GetUrl(), ref.databaseRecord.DatabaseName,
		ref.replicationFactor)
}

func (ref *CreateDatabaseOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	if resp == nil {
		err = errors.New("resp is invalid.")
	} else if resp.StatusCode == 201 {
		return Utils.ResponseToJSON(resp)
	} else if resp.StatusCode == 400 {
		//	судя по экзекутору, такой код сюда все равно не должен доходить
		return nil, errors.New("Error statusCode")
	}

	return
}

//# --------------------DeleteDatabaseOperation------------------------------
type tDeleteParameters struct {
	DatabaseNames             string
	HardDelete                bool
	FromNodes                 SrvNodes.IServerNode
	TimeToWaitForConfirmation time.Duration
}
type DeleteDatabaseOperation struct {
	ServerOperation
	parameters *tDeleteParameters
}

func NewDeleteDatabaseOperation(dbName string, hardDelete bool, fromNode SrvNodes.IServerNode, toWaitForConfirmation time.Duration) (ref *DeleteDatabaseOperation, err error) {
	err = Utils.DatabaseNameValidation(dbName)
	if err != nil {
		return nil, err
	}

	ref = &DeleteDatabaseOperation{}
	ref.init()
	ref.Method = "DELETE"
	ref.parameters = &tDeleteParameters{
		DatabaseNames:             dbName,
		HardDelete:                hardDelete,
		TimeToWaitForConfirmation: toWaitForConfirmation,
		FromNodes:                 fromNode,
	}

	ref.Data = ref.parameters

	return
}

func (ref *DeleteDatabaseOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/admin/databases", serverNode.GetUrl())
}

func (ref *DeleteDatabaseOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	//try:
	//	response = response.json()
	//	if "Error" in
	//response:
	//	raise
	//	exceptions.DatabaseDoesNotExistException(response["Message"])
	//	except
	//return nil, errors.New:
	//	raise
	//	response.raise_for_status()
	//	return
	//	{
	//		"raft_command_index": response["RaftCommandIndex"]
	//	}
	return
}

//# --------------------CertificateOperation------------------------------
type GetDatabaseNamesOperation struct {
	ServerOperation
	start, pageSize int
}

func NewGetDatabaseNamesOperation(start, pageSize int) *GetDatabaseNamesOperation {
	ref := &GetDatabaseNamesOperation{}
	ref.init()
	ref.Method = "GET"
	ref.start = start
	ref.pageSize = pageSize

	return ref
}
func (ref *GetDatabaseNamesOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/databases?start=%d&pageSize=%d&namesOnly=true", serverNode.GetUrl(), ref.start,
		ref.pageSize)
}
func (ref *GetDatabaseNamesOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {

	if resp == nil {
		err = errors.New("resp is invalid.")
	} else if resp.StatusCode == 201 {
		return Utils.ResponseToJSON(resp)
	} else if resp.StatusCode == 400 {
		//	судя по экзекутору, такой код сюда все равно не должен доходить
		return nil, errors.New("Error statusCode")
	}

	return
	//	if response is
	//None:
	//	raise
	//	return nil, errors.New("Invalid response")
	//
	//	response = response.json()
	//	if "Error" in
	//response:
	//	raise
	//	exceptions.ErrorResponseException(response["Error"])
	//
	//	if "Databases" not
	//	in
	//response:
	//	raise
	//	return nil, errors.New("Invalid response")
	//
	//	return response["Databases"]
}

//# --------------------GetDatabaseRecordOperation------------------------------
type GetDatabaseRecordOperation struct {
	ServerOperation
	databaseName string
}

func NewGetDatabaseRecordOperation(dbName string) (ref *GetDatabaseRecordOperation) {
	ref = &GetDatabaseRecordOperation{}
	ref.init()
	ref.Method = "GET"
	ref.databaseName = dbName

	return
}
func (ref *GetDatabaseRecordOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/admin/databases?name=%d", serverNode.GetUrl(), ref.databaseName)
}
func (ref *GetDatabaseRecordOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	//try:
	//response = response.json()
	//if "Error" in response:
	//raise exceptions.ErrorResponseException(response["Error"])
	//return response["Topology"]
	//except:
	//raise response.raise_for_status(
	return
}

//# --------------------CertificateOperation------------------------------

type GetCertificateOperation struct {
	ServerOperation
	start, pageSize int
}

func NewGetCertificateOperation(start, pageSize int) *GetCertificateOperation {
	ref := &GetCertificateOperation{}
	ref.init()
	ref.Method = "GET"
	ref.start = start
	ref.pageSize = pageSize

	return ref
}

func (ref *GetCertificateOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/databases?start=%d&pageSize=%d", serverNode.GetUrl(), ref.start,
		ref.pageSize)
}
func (ref *GetCertificateOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	//if response is None:
	//raise return nil, errors.New("response is invalid.")
	//data = {}
	//try:
	//response = response.json()["Results"]
	//if len(response) > 1:
	//raise return nil, errors.New("response is Invalid")
	//for key, value in response[0].items() {
	//data[Utils.convert_to_snake_case(key)] = value
	//return Certificatefunc (ref *Operationinition(**data)
	//
	//except return nil, errors.New:
	//raise response.raise_for_status()
	//	}
	return
}

type tCertificationPermission map[string]string
type CreateClientCertificateOperation struct {
	ServerOperation
	name, clearance, password string
	permissions               tCertificationPermission
}

//Add certificate json to the server and get thumbprint from server to use
//@param str name: The name of the certificate
//@param Dict[str:DatabaseAccess] permissions: the permissions to the database the key is the name of the database
//the value is database access (read or admin)
//@param SecurityClearance clearance: The clearance of the client
//@param str password: The password of the certificate
func NewCreateClientCertificateOperation(name string, permissions tCertificationPermission, clearance, password string) (*CreateClientCertificateOperation, error) {
	if name == "" {
		return nil, errors.New("name cannot by None")
	}
	if (permissions == nil) || (len(permissions) == 0) {
		return nil, errors.New("permissions cannot be None")
	}
	ref := &CreateClientCertificateOperation{}
	ref.init()
	ref.Method = "POST"
	ref.IsReadRequest = true
	ref.UseStream = true
	ref.name = name
	ref.permissions = permissions
	if clearance > "" {
		ref.clearance = clearance
	} else {
		ref.clearance = SecurityClearance.UnauthenticatedClients
	}

	ref.password = password

	return ref, nil
}

func (ref *CreateClientCertificateOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = serverNode.GetUrl() + "/admin/certificates"
	var data = map[string]interface{}{"Name": ref.name, "SecurityClearance": ref.clearance}
	if ref.password > "" {
		data["Password"] = ref.password
	}
	data["Permissions"] = ref.permissions
	ref.Data = data
}
func (ref *CreateClientCertificateOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	//	if response is
	//None:
	//	return None
	//
	//	if response.status_code != 201:
	//	response = response.json()
	//	if "Error" in
	//response:
	//	raise
	//	exceptions.ErrorResponseException(response["Error"])
	//	return response.raw.data
	return
}

type PutClientCertificateOperation struct {
	ServerOperation
	name        string
	certificate io.ByteReader
	permissions tCertificationPermission
	clearance   string
}

//@param str name: Certificate name
//@param x509 certificate: X509 thumbprint file (OpenSSL.crypto.load_certificate(OpenSSL.crypto.FILETYPE_PEM, pem))
//@param Dict[str:DatabaseAccess] permissions: the permissions to the client
//@param SecurityClearance clearance: The clearance of the client
func NewPutClientCertificateOperation(name string, certificate io.ByteReader, permissions tCertificationPermission, clearance string) (*PutClientCertificateOperation, error) {
	if certificate == nil {
		return nil, errors.New("certificate cannot be None")
	}
	if name == "" {
		return nil, errors.New("name cannot by None")
	}
	if (permissions == nil) || (len(permissions) == 0) {
		return nil, errors.New("permissions cannot be None")
	}
	ref := &PutClientCertificateOperation{}
	ref.init()
	ref.Method = "POST"
	ref.UseStream = true
	ref.name = name
	ref.certificate = certificate
	ref.permissions = permissions
	if clearance > "" {
		ref.clearance = clearance
	} else {
		ref.clearance = SecurityClearance.UnauthenticatedClients
	}

	return ref, nil
}

func (ref *PutClientCertificateOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = serverNode.GetUrl() + "/admin/certificates"

	ref.Data = map[string]interface{}{"Name": ref.name, "Certificate": ref.certificate, "SecurityClearance": ref.clearance, "Permissions": ref.permissions}
}

type DeleteCertificateOperation struct {
	ServerOperation
	thumbprint string
}

// NewDeleteCertificateOperation -
func NewDeleteCertificateOperation(thumbprint string) (*DeleteCertificateOperation, error) {
	if thumbprint == "" {
		return nil, errors.New("thumbprint cannot be None")
	}
	ref := &DeleteCertificateOperation{}
	ref.init()
	ref.Method = "DELETE"
	ref.thumbprint = thumbprint

	return ref, nil
}

func (ref *DeleteCertificateOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/admin/certificates?thumbprint=%s", serverNode.GetUrl(), strconv.Quote(ref.thumbprint))
}
