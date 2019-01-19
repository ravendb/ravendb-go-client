package ravendb

// TcpConnectionInfo describes tpc connection
type TcpConnectionInfo struct {
	Port        int     `json:"Port"`
	URL         string  `json:"Url"`
	Certificate *string `json:"Certificate"`
}
