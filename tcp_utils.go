package ravendb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
)

func newTLSConfig(certificate *tls.Certificate, trustStore *x509.Certificate) (*tls.Config, error) {
	if certificate != nil && trustStore == nil {
		return nil, newIllegalArgumentError("certificates and trustStoreASN1 can't be both empty")
	}

	config := &tls.Config{}

	if trustStore != nil {
		roots := x509.NewCertPool()
		roots.AddCert(trustStore)
		config.RootCAs = roots
	}
	// TODO: not sure if this should always (ever?) be set
	// see setSSLHostnameVerifier and loadTrustMaterial in java code
	config.InsecureSkipVerify = true

	config.Certificates = []tls.Certificate{*certificate}
	return config, nil
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
		// This is a root signing certificate needed for tls.Dial to recognize
		// data send by the server (?) as properly signed.
		// If we didn't have this we could set tls.Config.InsecureSkipVerify to true
		var trustStore *x509.Certificate
		if len(serverCertificateBase64) > 0 {
			serverCertificate, err := base64.StdEncoding.DecodeString(string(serverCertificateBase64))
			if err != nil {
				fmt.Printf("base64.StdEncoding.DecodeString() failed with %s\n", err)
				return nil, err
			}
			trustStore, err = x509.ParseCertificate(serverCertificate)
			if err != nil {
				return nil, err
			}
		}

		config, err := newTLSConfig(clientCertificate, trustStore)
		if err != nil {
			return nil, err
		}
		// forcing TLS 1.2 as Java code seems to be doing
		config.MinVersion = tls.VersionTLS12
		config.MaxVersion = tls.VersionTLS12

		conn, err := tls.Dial("tcp", parsed.Host, config)
		if err != nil {
			fmt.Printf("tls.Dial() failed with %s\n", err)
		}
		return conn, err
	}

	// parsed.Host is in the form "127.0.0.1:14206"
	return net.Dial("tcp", parsed.Host)
}
