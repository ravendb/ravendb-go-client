package ravendb

import (
	"time"
)

type Operation struct {
	_requestExecutor *RequestExecutor
	//TBD private readonly Func<DatabaseChanges> _changes;
	_conventions *DocumentConventions
	_id          int

	// if true, this represents ServerWideOperation
	IsServerWide bool
}

func (o *Operation) GetID() int {
	return o._id
}

func NewOperation(requestExecutor *RequestExecutor, changes func() *DatabaseChanges, conventions *DocumentConventions, id int) *Operation {
	return &Operation{
		_requestExecutor: requestExecutor,
		//TBD _changes = changes;
		_conventions: conventions,
		_id:          id,
	}
}

func (o *Operation) fetchOperationsStatus() (ObjectNode, error) {
	command := o.getOperationStateCommand(o._conventions, o._id)
	err := o._requestExecutor.ExecuteCommand(command)
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

func (o *Operation) getOperationStateCommand(conventions *DocumentConventions, id int) RavenCommand {
	if o.IsServerWide {
		return NewGetServerWideOperationStateCommand(o._conventions, id)
	}
	return NewGetOperationStateCommand(o._conventions, o._id)
}

func (o *Operation) WaitForCompletion() error {
	for {
		status, err := o.fetchOperationsStatus()
		if err != nil {
			return err
		}

		operationStatus, ok := JsonGetAsText(status, "Status")
		if !ok {
			return newRavenError("missing 'Status' field in response")
		}
		switch operationStatus {
		case "Completed":
			return nil
		case "Cancelled":
			return newOperationCancelledError("")
		case "Faulted":
			result, ok := status["Result"].(ObjectNode)
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
