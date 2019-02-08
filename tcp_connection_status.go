package ravendb

type tcpConnectionStatus = string

const (
	tcpConnectionStatusOk                  = "Ok"
	tcpConnectionStatusAuthorizationFailed = "AuthorizationFailed"
	tcpConnectionStatusTcpVersionMismatch  = "TcpVersionMismatch"
)
