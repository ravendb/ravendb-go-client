package ravendb

type TcpConnectionInfo struct {
	Port        int     `json:"Port"`
	Url         string  `json:"Url"`
	Certificate *string `json:"Certificate"`
}

func (i *TcpConnectionInfo) getPort() int {
	return i.Port
}

func (i *TcpConnectionInfo) getUrl() String {
	return i.Url
}

func (i *TcpConnectionInfo) getCertificate() *string {
	return i.Certificate
}

/*
public void setPort(int port) {
	this.port = port;
}

public void setUrl(String url) {
	this.url = url;
}

public void setCertificate(String certificate) {
	this.certificate = certificate;
}
*/
