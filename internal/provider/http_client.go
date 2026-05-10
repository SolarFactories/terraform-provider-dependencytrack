package provider

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"

	dtrack "github.com/DependencyTrack/client-go"
)

type (
	Header struct {
		Name  string
		Value string
	}

	transport struct {
		inner   http.RoundTripper
		headers []Header
	}
)

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, header := range t.headers {
		req.Header.Add(header.Name, header.Value)
	}
	return t.inner.RoundTrip(req)
}

func NewHTTPClient(headers []Header, pemCerts []byte, clientCertFile string, clientKeyFile string) (*http.Client, error) {
	// Create x509.CertPool for RootCA.
	rootCAs, err := newCertPool(pemCerts)
	if err != nil {
		return nil, err
	}
	// Create and configure underlying transport for TLS.
	innerTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("expected http.DefaultTransport to be a *http.Transport. Found %T", http.DefaultTransport)
	}
	innerTransport.TLSClientConfig = &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS13,
	}
	// Configure mTLS.
	if clientCertFile != "" && clientKeyFile != "" {
		innerTransport.TLSClientConfig.MinVersion = tls.VersionTLS13
		keypair, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return nil, err
		}
		innerTransport.TLSClientConfig.Certificates = []tls.Certificate{keypair}
	}
	return &http.Client{
		Timeout: dtrack.DefaultTimeout,
		Transport: &transport{
			inner:   innerTransport,
			headers: headers,
		},
	}, nil
}

func newCertPool(pemCerts []byte) (*x509.CertPool, error) {
	if len(pemCerts) == 0 {
		return x509.SystemCertPool()
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemCerts) {
		return nil, errors.New("invalid PEM certificates used for root ca")
	}
	return certPool, nil
}
