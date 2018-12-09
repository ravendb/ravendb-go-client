package ravendb

type ILazyOperation interface {
	createRequest() *GetRequest
	getResult() interface{}
	getQueryResult() *QueryResult
	isRequiresRetry() bool
	handleResponse(response *GetResponse) error
}
