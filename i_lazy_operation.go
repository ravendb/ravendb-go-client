package ravendb

// ILazyOperation defines methods required to implement lazy operation
type ILazyOperation interface {
	createRequest() *GetRequest
	getResult() interface{}
	getQueryResult() *QueryResult
	isRequiresRetry() bool
	handleResponse(response *GetResponse) error
}
