package raven_operations

import (
	"testing"
	//httpSelf "github.com/ravendb/ravendb-go-client/http"
	//"github.com/ravendb/ravendb-go-client/http/commands"
	SrvNodes "github.com/ravendb/ravendb-go-client/http/server_nodes"
)
func TestAdminOperation_GetOperation(t *testing.T) {
	oper := &AdminOperation{}
	oper.init()
	if oper.GetOperation() != "AdminOperation" {
		t.Errorf("not valid AdminOperation")
	}
}
func TestDeleteIndexOperation_CreateRequest(t *testing.T) {
	const validUrl = `/databases/test/indexes?name=["test_index"]`
	var srvNode serverNode

	srvNode.database = "test"
	_, err := NewDeleteIndexOperation("")
	if err == nil {
		t.Errorf("error in create DeleteIndexOperation")
	}
	oper, err := NewDeleteIndexOperation("test_index")
	oper.CreateRequest(srvNode)

	if oper.Url != validUrl {
		t.Errorf("not valid URL. Wait %s, answer - %s", validUrl, oper.Url)
	}
}
func TestNewCreateDatabaseOperation(t *testing.T) {

}
func TestCreateDatabaseOperation_CreateRequest(t *testing.T) {
	const validUrl = "/admin/databases?name=test&replication-factor=1"
	oper, err := NewCreateDatabaseOperation("test", 0, nil, nil)
	if err != nil {
		t.Fail()
	}
	var srvNode SrvNodes.ServerNode
	oper.CreateRequest(srvNode)

	if oper.Url != validUrl {
		t.Errorf("not valid URL. Wait %s, answer - %s", validUrl, oper.Url)
	}

	if command, err := oper.GetCommand(""); (err == nil) || (command != nil) {
		t.Errorf("not valid answer GetCommand if conv is empty")
	}

	if command, err := oper.GetCommand("first"); (err != nil) || (command == nil) {
		t.Errorf("not valid answer GetCommand if conv is not empty")
	} else {
		tCom := command.GetHeaders()
		t.Log(tCom)
	}

	return
}
//func TestPut(t *testing.T)  {
//
//	requestExecutor, _ := httpSelf.CreateForSingleNode("http://localhost:8080", "test")
//	//putCommand, _ := commands.NewPutDocumentCommand("testing/1", testing2.Product{Name:"test"})
//	getCommand, _ := commands.NewGetDocumentCommand("testing/1", nil, false)
//	//requestExecutor.ExecuteOnCurrentNode(putCommand, false)
//	resp, _ := requestExecutor.ExecuteOnCurrentNode(getCommand, false)
//	t.Log(resp)
//	//var prod testing2.Product
//	//json.Unmarshal(resp, prod)
//	//if prod.Name != "test"{
//	//}
//}
