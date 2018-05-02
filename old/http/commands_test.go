package http

import (
	"encoding/json"
	"testing"

	"github.com/ravendb/ravendb-go-client/http/commands"
	testing2 "github.com/ravendb/ravendb-go-client/testing"
)

func TestPut(t *testing.T) {
	requestExecutor, _ := CreateForSingleNode("http://localhost:8080", "test")
	putCommand, _ := commands.NewPutDocumentCommand("testing/1", testing2.Product{Name: "test"})
	getCommand, _ := commands.NewGetDocumentCommand("testing/1", nil, false)
	requestExecutor.ExecuteOnCurrentNode(putCommand, false)
	resp, _ := requestExecutor.ExecuteOnCurrentNode(getCommand, false)
	var prod testing2.Product
	json.Unmarshal(resp, prod)
	if prod.Name != "test" {
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	putCommand, _ := commands.NewPutDocumentCommand("products/101", testing2.Product{Name: "test"})
	putCommand2, _ := commands.NewPutDocumentCommand("products/10", testing2.Product{Name: "test"})
	getCommand, _ := commands.NewGetDocumentCommand("products/101", nil, false)
	getCommand2, _ := commands.NewGetDocumentCommand("products/10", nil, false)

	requestExecutorPtr, _ := CreateForSingleNode("http://localhost:8080", "test")
	requestExecutorPtr.ExecuteOnCurrentNode(putCommand, false)
	requestExecutorPtr.ExecuteOnCurrentNode(putCommand2, false)

	response, _ := requestExecutorPtr.ExecuteOnCurrentNode(getCommand, false)
	response2, _ := requestExecutorPtr.ExecuteOnCurrentNode(getCommand2, false)

	var prod testing2.Product
	var prod2 testing2.Product
	json.Unmarshal(response, prod)
	json.Unmarshal(response2, prod2)
	if prod.Name != "test" || prod2.Name != "test" {
		t.Fail()
	}
}
