package ravendb

// Note: for simplicity, ServerWideOperation is folded into Operation

func NewServerWideOperation(requestExecutor *RequestExecutor, conventions *DocumentConventions, id int64) *Operation {
	res := NewOperation(requestExecutor, nil, conventions, id)
	res.IsServerWide = true
	return res
}
