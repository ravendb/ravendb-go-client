package ravendb

import (
	"time"
)

// Operation describes async operation being executed on the server
type Operation struct {
	requestExecutor *RequestExecutor
	//TBD private readonly Func<databaseChanges> _changes;
	conventions *DocumentConventions
	id          int64

	// if true, this represents ServerWideOperation
	IsServerWide bool
}

func (o *Operation) GetID() int64 {
	return o.id
}

func NewOperation(requestExecutor *RequestExecutor, changes func() *databaseChanges, conventions *DocumentConventions, id int64) *Operation {
	return &Operation{
		requestExecutor: requestExecutor,
		//TBD _changes = changes;
		conventions: conventions,
		id:          id,
	}
}

func (o *Operation) fetchOperationsStatus() (map[string]interface{}, error) {
	command := o.getOperationStateCommand(o.conventions, o.id)
	err := o.requestExecutor.ExecuteCommand(command, nil)
	if err != nil {
		return nil, err
	}

	switch cmd := command.(type) {
	case *GetOperationStateCommand:
		return cmd.Result, nil
	case *GetServerWideOperationStateCommand:
		return cmd.Result, nil
	}
	panicIf(true, "Unexpected command type %T", command)
	return nil, nil
}

func (o *Operation) getOperationStateCommand(conventions *DocumentConventions, id int64) RavenCommand {
	if o.IsServerWide {
		return NewGetServerWideOperationStateCommand(o.conventions, id)
	}
	return NewGetOperationStateCommand(o.conventions, o.id)
}

func (o *Operation) WaitForCompletion() error {
	for {
		status, err := o.fetchOperationsStatus()
		if err != nil {
			return err
		}

		operationStatus, ok := jsonGetAsText(status, "Status")
		if !ok {
			return newRavenError("missing 'Status' field in response")
		}
		switch operationStatus {
		case "Completed":
			return nil
		case "Cancelled":
			return newOperationCancelledError("")
		case "Faulted":
			result, ok := status["Result"].(map[string]interface{})
			if !ok {
				return newRavenError("status has no 'Result' object. Status: #%v", status)
			}
			var exceptionResult OperationExceptionResult
			err = structFromJSONMap(result, &exceptionResult)
			if err != nil {
				return err
			}
			return exceptionDispatcherGet2(exceptionResult.Message, exceptionResult.Error, exceptionResult.Type, exceptionResult.StatusCode)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
