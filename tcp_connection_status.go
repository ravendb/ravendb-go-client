package ravendb

type TcpConnectionStatus = string

const (
	TcpConnectionStatusOk                  = "Ok"
	TcpConnectionStatusAuthorizationFailed = "AuthorizationFailed"
	TcpConnectionStatusTcpVersionMismatch  = "TcpVersionMismatch"
)
