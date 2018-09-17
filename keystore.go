package ravendb

import "crypto/tls"

// KeyStore helps porting Java code
type KeyStore struct {
	Certificates []tls.Certificate
}
