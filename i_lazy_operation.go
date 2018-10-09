package ravendb

type ILazyOperation interface {
	createRequest() *GetRequest
	getResult() Object
	getQueryResult() *QueryResult
	isRequiresRetry() bool
	handleResponse(response *GetResponse) error
}
