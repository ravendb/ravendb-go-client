package ravendb

import "crypto/tls"

// KeyStoreCertificate represents a single certificate in KeyStore
type KeyStoreCertificate struct {
	TLSCert *tls.Certificate
	PEM     []byte // raw PEM data of the certificate
}

// KeyStore helps porting Java code
type KeyStore struct {
	Certificates []*KeyStoreCertificate
}
