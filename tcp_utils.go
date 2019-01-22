package ravendb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
)

func tcpConnect(uri string, serverCertificateBase64 []byte, clientCertificate *KeyStore) (net.Conn, error) {
	//  uri is in the format: tcp://127.0.0.1:14206
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "tcp" {
		return nil, fmt.Errorf("bad url: '%s', expected scheme to be 'ftp', is '%s'", uri, parsed.Scheme)
	}

	if len(serverCertificateBase64) > 0 && clientCertificate != nil {
		// serverCertificateBase64 is base64-encoded ASN1-encoded certificate
		// This is a root signing certificate needed for tls.Dial to recognize
		// data send by the server (?) as properly signed.
		// If we didn't have this we could set tls.Config.InsecureSkipVerify to true
		serverCertificate, err := base64.StdEncoding.DecodeString(string(serverCertificateBase64))
		if err != nil {
			fmt.Printf("base64.StdEncoding.DecodeString() failed with %s\n", err)
			return nil, err
		}
		roots := x509.NewCertPool()

		cert, err := x509.ParseCertificate(serverCertificate)
		if err != nil {
			return nil, err
		}
		roots.AddCert(cert)

		config := &tls.Config{
			RootCAs: roots,
			// forcing TLS 1.2 as Java code seems to be doing
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12,
		}
		for _, cert := range clientCertificate.Certificates {
			config.Certificates = append(config.Certificates, *cert.TLSCert)
		}

		conn, err := tls.Dial("tcp", parsed.Host, config)
		if err != nil {
			fmt.Printf("tls.Dial() failed with %s\n", err)
		}
		return conn, err
	}

	// parsed.Host is in the form "127.0.0.1:14206"
	return net.Dial("tcp", parsed.Host)
}
