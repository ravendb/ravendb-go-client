package ravendb

// TcpConnectionHeaderResponse describes tcp connection header response
type TcpConnectionHeaderResponse struct {
	Status  TcpConnectionStatus `json:"Status"`
	Message string              `json:"Message"`
	Version int                 `json:"Version"`
}
