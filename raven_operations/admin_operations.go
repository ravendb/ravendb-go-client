package raven_operations

import (
	"errors"
	"github.com/ravendb-go-client/http/commands"
	"fmt"
	"strconv"
	"net/http"

)
// AdminOperation abstract class  - root operations classes
type AdminOperation struct {
	command *commands.Command
	operation string
	Url string
}
func (ref *AdminOperation) init()  {
	ref.operation = "AdminOperation"
	ref.command, _ = commands.NewRavenCommand()
	//ref.command.Init()

}
func (obj AdminOperation) GetOperation() string{
	return obj.operation
}
func (ref *AdminOperation) GetCommand(conventions string) (*AdminOperation, error) {
	return ref, nil
}
func (ref *AdminOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	return nil, nil
}
// serverNode - unknow class
type serverNode struct {
	url, database   string
}
func (serverNode serverNode) getFormatURL(suffux string, args ... interface{}) string {
	values := [] interface{} {serverNode.url, serverNode.database}
	values = append(values, args)
	return fmt.Sprintf("%s/databases/%s/" + suffux, values ...)
}
// begin declaration DeleteIndexOperation class
type DeleteIndexOperation struct {
	AdminOperation
	indexName string
	resp *http.Response
}
// constructor DeleteIndexOperation (must have not null indexName)
func NewDeleteIndexOperation(indexName string) (ref *DeleteIndexOperation, err error) {
	if indexName == "" {
		return nil, errors.New("Invalid indexName")
	}
	ref = &DeleteIndexOperation{}
	ref.init()
	ref.indexName = indexName
	ref.command.Method = "DELETE"

	return
}
func (obj *DeleteIndexOperation) CreateRequest(serverNode serverNode) {
	obj.Url = serverNode.getFormatURL("indexes?name=%s", strconv.Quote(obj.indexName))
}
func (obj *DeleteIndexOperation) GetResponseRaw(resp *http.Response) {
	obj.resp = resp
}
// begin declaration GetIndexOperation class
type GetIndexOperation struct {
	AdminOperation
	indexName string
	resp *http.Response
}
// constructor GetIndexOperation (must have not null indexName)
func NewGetIndexOperation(indexName string) (ref *GetIndexOperation, err error) {
	if indexName == "" {
		return nil, errors.New("Invalid indexName")
	}
	ref = &GetIndexOperation{}
	ref.init()
	ref.indexName = indexName
	ref.command.Method = "GET"

	return
}
func (obj *GetIndexOperation) CreateRequest(serverNode serverNode) {
	obj.Url = serverNode.getFormatURL("indexes?name=%s", strconv.Quote(obj.indexName))
}
func (obj *GetIndexOperation) GetResponseRaw(resp *http.Response) {

	if resp == nil {
		return
	}
//	data = {}
//try:
//	response = response.json()["Results"]
//	if len(response) > 1:
//	raise ValueError("response is Invalid")
//	for key, value in response[0].items():
//	data[Utils.convert_to_snake_case(key)] = value
//	return IndexDefinition(**data)
//
//	except ValueError:
//	raise response.raise_for_status()
	}

// begin declaration PutIndexesCommand class
type PutIndexesOperation struct {
	AdminOperation
	indexName string
	resp *http.Response
}
// constructor PutIndexesOperation (must have not null indexName)
func NewPutIndexesOperation(indexName string) (ref *PutIndexesOperation, err error) {
	if indexName == "" {
		return nil, errors.New("Invalid indexName")
	}
	ref = &PutIndexesOperation{}
	ref.init()
	ref.indexName = indexName
	ref.command.Method = "DELETE"

	return
}
func (obj *PutIndexesOperation) CreateRequest(serverNode serverNode) {
	obj.Url = fmt.Sprintf("indexes?name=%s", strconv.Quote(obj.indexName))
}
func (obj *PutIndexesOperation) GetResponseRaw(resp *http.Response) {
	obj.resp = resp
}

