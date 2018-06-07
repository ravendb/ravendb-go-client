package ravendb

import "time"

type Operation struct {
	_requestExecutor *RequestExecutor
	//TBD private readonly Func<IDatabaseChanges> _changes;
	_conventions *DocumentConventions
	_id          int
}

func (o *Operation) getId() int {
	return o._id
}

func NewOperation(requestExecutor *RequestExecutor, changes *IDatabaseChanges, conventions *DocumentConventions, id int) *Operation {
	return &Operation{
		_requestExecutor: requestExecutor,
		//TBD _changes = changes;
		_conventions: conventions,
		_id:          id,
	}
}

func (o *Operation) fetchOperationsStatus() ObjectNode {
	command := o.getOperationStateCommand(o._conventions, o._id)
	o._requestExecutor.executeCommand(command)
	resi := command.getResult()
	res := resi.(ObjectNode)
	return res
}

func (o *Operation) getOperationStateCommand(conventions *DocumentConventions, id int) *RavenCommand {
	return NewGetOperationStateCommand(o._conventions, o._id)
}

func (o *Operation) waitForCompletion() error {
	for {
		status := o.fetchOperationsStatus()

		operationStatus := jsonGetAsText(status, "Status")
		switch operationStatus {
		case "Completed":
			return nil
		case "Cancelled":
			return NewOperationCancelledException("")
		case "Faulted":
			panicIf(true, "NYI")
			/*
				result := status["Result"]

				OperationExceptionResult exceptionResult = JsonExtensions.getDefaultMapper().convertValue(result, OperationExceptionResult.class);

				throw ExceptionDispatcher.get(exceptionResult.getMessage(), exceptionResult.getError(), exceptionResult.getType(), exceptionResult.getStatusCode());
			*/
		}

		time.Sleep(500 * time.Millisecond)
	}
}
