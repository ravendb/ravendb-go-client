package ravendb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
)

func newTLSConfig(certificate *tls.Certificate, trustStore *x509.Certificate) *tls.Config {
	config := &tls.Config{}

	if trustStore != nil {
		roots := x509.NewCertPool()
		roots.AddCert(trustStore)
		config.RootCAs = roots
	}

	if certificate != nil {
		config.Certificates = []tls.Certificate{*certificate}
	}
	return config
}

func tcpConnect(uri string, serverCertificateBase64 []byte, clientCertificate *tls.Certificate) (net.Conn, error) {
	//  uri is in the format: tcp://127.0.0.1:14206
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "tcp" {
		return nil, fmt.Errorf("bad url: '%s', expected scheme to be 'ftp', is '%s'", uri, parsed.Scheme)
	}

	if len(serverCertificateBase64) > 0 || clientCertificate != nil {
		// serverCertificateBase64 is base64-encoded ASN1-encoded certificate
		// The server certificate is supplied as the trust anchor for verifying the TCP TLS peer.
		var trustStore *x509.Certificate
		if len(serverCertificateBase64) > 0 {
			serverCertificate, err := base64.StdEncoding.DecodeString(string(serverCertificateBase64))
			if err != nil {
				return nil, err
			}
			trustStore, err = x509.ParseCertificate(serverCertificate)
			if err != nil {
				return nil, err
			}
		}

		config := newTLSConfig(clientCertificate, trustStore)
		// forcing TLS 1.2 as Java code seems to be doing
		config.MinVersion = tls.VersionTLS12
		config.MaxVersion = tls.VersionTLS12

		conn, err := tls.Dial("tcp", parsed.Host, config)
		return conn, err
	}

	// parsed.Host is in the form "127.0.0.1:14206"
	return net.Dial("tcp", parsed.Host)
}
