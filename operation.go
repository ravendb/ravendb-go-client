package ravendb

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
