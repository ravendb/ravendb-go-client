package ravendb

type TcpConnectionInfo struct {
	Port        int     `json:"Port"`
	Url         string  `json:"Url"`
	Certificate *string `json:"Certificate"`
}

func (i *TcpConnectionInfo) getPort() int {
	return i.Port
}

func (i *TcpConnectionInfo) getUrl() string {
	return i.Url
}

func (i *TcpConnectionInfo) getCertificate() *string {
	return i.Certificate
}

/*
public void setPort(int port) {
	this.port = port;
}

public void setUrl(string url) {
	this.url = url;
}

public void setCertificate(string certificate) {
	this.certificate = certificate;
}
*/
