package ravendb

type tcpConnectionHeaderResponse struct {
	Status  tcpConnectionStatus `json:"Status"`
	Message string              `json:"Message"`
	Version int                 `json:"Version"`
}
