package ravendb

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

type tlsTestAuthority struct {
	certificate *x509.Certificate
	privateKey  *ecdsa.PrivateKey
}

func newTLSTestAuthority(t *testing.T, name string) tlsTestAuthority {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal("cannot generate test CA key")
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal("cannot generate test CA serial")
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal("cannot create test CA certificate")
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal("cannot parse test CA certificate")
	}
	return tlsTestAuthority{certificate: certificate, privateKey: privateKey}
}

func (authority tlsTestAuthority) issue(t *testing.T, commonName string, dnsNames []string, usages ...x509.ExtKeyUsage) tls.Certificate {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal("cannot generate test certificate key")
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal("cannot generate test certificate serial")
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  usages,
		DNSNames:     dnsNames,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, authority.certificate, &privateKey.PublicKey, authority.privateKey)
	if err != nil {
		t.Fatal("cannot create test certificate")
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal("cannot parse test certificate")
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: privateKey, Leaf: leaf}
}

func startTLSHTTPTestServer(t *testing.T, serverCertificate tls.Certificate, clientAuthority *x509.Certificate, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewUnstartedServer(handler)
	config := &tls.Config{Certificates: []tls.Certificate{serverCertificate}, MinVersion: tls.VersionTLS12}
	if clientAuthority != nil {
		clientRoots := x509.NewCertPool()
		clientRoots.AddCert(clientAuthority)
		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = clientRoots
	}
	server.TLS = config
	server.StartTLS()
	t.Cleanup(server.Close)
	return server
}

func testURLWithHost(t *testing.T, server *httptest.Server, host string) string {
	t.Helper()
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal("cannot parse local test URL")
	}
	parsed.Host = net.JoinHostPort(host, parsed.Port())
	return parsed.String()
}

func testHTTPClient(t *testing.T, endpoint string, certificate *tls.Certificate, trustStore *x509.Certificate) *http.Client {
	t.Helper()
	executor := NewRequestExecutor("tls-test", certificate, trustStore, NewDocumentConventions(), []string{endpoint})
	t.Cleanup(executor.Close)
	client, err := executor.GetHTTPClient()
	if err != nil {
		t.Fatal("cannot create test request client")
	}
	return client
}

func doTLSRequest(client *http.Client, endpoint string) error {
	response, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return errUnexpectedTLSStatus
	}
	return nil
}

var errUnexpectedTLSStatus = &tlsTestError{}

type tlsTestError struct{}

func (*tlsTestError) Error() string { return "unexpected TLS fixture response" }

func TestHTTPServerCertificateVerification(t *testing.T) {
	trustedAuthority := newTLSTestAuthority(t, "trusted HTTP test CA")
	untrustedAuthority := newTLSTestAuthority(t, "untrusted HTTP test CA")
	serverCertificate := trustedAuthority.issue(t, "localhost server", []string{"localhost"}, x509.ExtKeyUsageServerAuth)
	server := startTLSHTTPTestServer(t, serverCertificate, nil, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := doTLSRequest(testHTTPClient(t, testURLWithHost(t, server, "localhost"), nil, trustedAuthority.certificate), testURLWithHost(t, server, "localhost")); err != nil {
		t.Fatal("valid HTTP server certificate and hostname were rejected")
	}
	if err := doTLSRequest(testHTTPClient(t, testURLWithHost(t, server, "localhost"), nil, untrustedAuthority.certificate), testURLWithHost(t, server, "localhost")); err == nil {
		t.Fatal("HTTP accepted a server certificate from an untrusted CA")
	}
	if err := doTLSRequest(testHTTPClient(t, testURLWithHost(t, server, "127.0.0.1"), nil, trustedAuthority.certificate), testURLWithHost(t, server, "127.0.0.1")); err == nil {
		t.Fatal("HTTP accepted a hostname mismatch")
	}
	clientCertificate := untrustedAuthority.issue(t, "system-root fallback client", nil, x509.ExtKeyUsageClientAuth)
	if err := doTLSRequest(testHTTPClient(t, server.URL, &clientCertificate, nil), server.URL); err == nil {
		t.Fatal("HTTP accepted a private test CA when only system roots were configured")
	}
	if err := doTLSRequest(testHTTPClient(t, server.URL, nil, nil), server.URL); err == nil {
		t.Fatal("HTTP accepted a private test CA through system roots")
	}
}

func TestHTTPClientCertificateAuthenticationAndOptionalCertificate(t *testing.T) {
	authority := newTLSTestAuthority(t, "HTTP mTLS test CA")
	serverCertificate := authority.issue(t, "localhost mTLS server", []string{"localhost"}, x509.ExtKeyUsageServerAuth)
	clientCertificate := authority.issue(t, "approved client", nil, x509.ExtKeyUsageClientAuth)
	server := startTLSHTTPTestServer(t, serverCertificate, authority.certificate, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.TLS == nil || len(request.TLS.PeerCertificates) != 1 || request.TLS.PeerCertificates[0].Subject.CommonName != "approved client" {
			http.Error(w, "client certificate identity mismatch", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	endpoint := testURLWithHost(t, server, "localhost")
	if err := doTLSRequest(testHTTPClient(t, endpoint, &clientCertificate, authority.certificate), endpoint); err != nil {
		t.Fatal("HTTP client-certificate authentication failed")
	}
	if err := doTLSRequest(testHTTPClient(t, endpoint, nil, authority.certificate), endpoint); err == nil {
		t.Fatal("HTTP server accepted a connection without its required client certificate")
	}
}

func TestHTTPExecutorsKeepIndependentTLSCredentials(t *testing.T) {
	authorityA := newTLSTestAuthority(t, "independent HTTP CA A")
	authorityB := newTLSTestAuthority(t, "independent HTTP CA B")
	serverA := startTLSHTTPTestServer(t, authorityA.issue(t, "localhost A", []string{"localhost"}, x509.ExtKeyUsageServerAuth), authorityA.certificate, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.TLS.PeerCertificates[0].Subject.CommonName != "client A" {
			http.Error(w, "client identity mismatch", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	serverB := startTLSHTTPTestServer(t, authorityB.issue(t, "localhost B", []string{"localhost"}, x509.ExtKeyUsageServerAuth), authorityB.certificate, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.TLS.PeerCertificates[0].Subject.CommonName != "client B" {
			http.Error(w, "client identity mismatch", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	clientA := testHTTPClient(t, testURLWithHost(t, serverA, "localhost"), certificatePtr(authorityA.issue(t, "client A", nil, x509.ExtKeyUsageClientAuth)), authorityA.certificate)
	clientB := testHTTPClient(t, testURLWithHost(t, serverB, "localhost"), certificatePtr(authorityB.issue(t, "client B", nil, x509.ExtKeyUsageClientAuth)), authorityB.certificate)
	endpointA := testURLWithHost(t, serverA, "localhost")
	endpointB := testURLWithHost(t, serverB, "localhost")
	if err := doTLSRequest(clientA, endpointA); err != nil {
		t.Fatal("first HTTP client lost its certificate or trust root")
	}
	if err := doTLSRequest(clientB, endpointB); err != nil {
		t.Fatal("second HTTP client lost its certificate or trust root")
	}
	if err := doTLSRequest(clientA, endpointB); err == nil {
		t.Fatal("first HTTP client unexpectedly used the second client's trust root")
	}
	if err := doTLSRequest(clientB, endpointA); err == nil {
		t.Fatal("second HTTP client unexpectedly used the first client's trust root")
	}
}

func certificatePtr(certificate tls.Certificate) *tls.Certificate { return &certificate }

func startTLSTCPTestServer(t *testing.T, serverCertificate tls.Certificate, clientAuthority *x509.Certificate) (string, <-chan error) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("cannot start TCP TLS fixture")
	}
	config := &tls.Config{Certificates: []tls.Certificate{serverCertificate}, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12}
	if clientAuthority != nil {
		clientRoots := x509.NewCertPool()
		clientRoots.AddCert(clientAuthority)
		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = clientRoots
	}
	server := tls.NewListener(listener, config)
	t.Cleanup(func() { _ = server.Close() })
	done := make(chan error, 1)
	go func() {
		connection, acceptErr := server.Accept()
		if acceptErr != nil {
			done <- acceptErr
			return
		}
		tlsConnection := connection.(*tls.Conn)
		handshakeErr := tlsConnection.Handshake()
		_ = tlsConnection.Close()
		done <- handshakeErr
	}()
	return listener.Addr().String(), done
}

func tcpTrustRoot(authority *x509.Certificate) []byte {
	return []byte(base64.StdEncoding.EncodeToString(authority.Raw))
}

func TestTCPServerCertificateVerificationAndClientAuthentication(t *testing.T) {
	authority := newTLSTestAuthority(t, "TCP test CA")
	untrustedAuthority := newTLSTestAuthority(t, "untrusted TCP test CA")
	serverCertificate := authority.issue(t, "localhost TCP server", []string{"localhost"}, x509.ExtKeyUsageServerAuth)
	clientCertificate := authority.issue(t, "approved TCP client", nil, x509.ExtKeyUsageClientAuth)

	address, handshake := startTLSTCPTestServer(t, serverCertificate, nil)
	connection, err := tcpConnect("tcp://localhost:"+strconv.Itoa(portFromAddress(t, address)), tcpTrustRoot(authority.certificate), nil)
	if err != nil {
		t.Fatal("TCP rejected a valid server certificate and hostname without a client certificate")
	}
	_ = connection.Close()
	if err := <-handshake; err != nil {
		t.Fatal("TCP server handshake failed")
	}

	address, handshake = startTLSTCPTestServer(t, serverCertificate, nil)
	connection, err = tcpConnect("tcp://localhost:"+strconv.Itoa(portFromAddress(t, address)), tcpTrustRoot(untrustedAuthority.certificate), nil)
	if err == nil {
		_ = connection.Close()
		t.Fatal("TCP accepted a server certificate from an untrusted CA")
	}
	<-handshake

	address, handshake = startTLSTCPTestServer(t, serverCertificate, nil)
	connection, err = tcpConnect("tcp://127.0.0.1:"+strconv.Itoa(portFromAddress(t, address)), tcpTrustRoot(authority.certificate), nil)
	if err == nil {
		_ = connection.Close()
		t.Fatal("TCP accepted a hostname mismatch")
	}
	<-handshake

	address, handshake = startTLSTCPTestServer(t, serverCertificate, nil)
	connection, err = tcpConnect("tcp://localhost:"+strconv.Itoa(portFromAddress(t, address)), nil, &clientCertificate)
	if err == nil {
		_ = connection.Close()
		t.Fatal("TCP accepted a private test CA when only system roots were configured")
	}
	<-handshake

	address, handshake = startTLSTCPTestServer(t, serverCertificate, authority.certificate)
	connection, err = tcpConnect("tcp://localhost:"+strconv.Itoa(portFromAddress(t, address)), tcpTrustRoot(authority.certificate), &clientCertificate)
	if err != nil {
		t.Fatal("TCP client-certificate authentication failed")
	}
	_ = connection.Close()
	if err := <-handshake; err != nil {
		t.Fatal("TCP client-certificate server handshake failed")
	}
}

func portFromAddress(t *testing.T, address string) int {
	t.Helper()
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal("cannot parse local TCP fixture address")
	}
	parsed, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal("cannot parse local TCP fixture port")
	}
	return parsed
}
