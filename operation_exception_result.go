package ravendb

// OperationExceptionResult represents an exception information from the server
type OperationExceptionResult struct {
	Type       string `json:"Type"`
	Message    string `json:"Message"`
	Error      string `json:"Error"`
	StatusCode int    `json:"StatusCode"`
}
