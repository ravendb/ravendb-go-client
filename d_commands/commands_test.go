package d_commands

//import (
//	"testing"
//)
//
//func TestPutSuccess(t *testing.T){
//	requestExecutor := NewRequestExecutor()
//	putCommand := NewPutDocumentCommand("testing/1", map[string]interface{}{"Name": "test","@metadata":{}})
//	requestExecutor.execute(putCommand)
//	response := requestExecutor.execute(NewGetDocumentCommand("testing/1"))
//	if response["Results"][0]["@metadata"]["@id"] != "testing/1"{
//		t.Fail()
//	}
//
//}