package provider

import (
	"github.com/DependencyTrack/client-go"
	"net/http"
)

type Header struct {
	Name  string
	Value string
}

type transport struct {
	inner   http.RoundTripper
	headers []Header
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, header := range t.headers {
		req.Header.Add(header.Name, header.Value)
	}
	return t.inner.RoundTrip(req)
}

func NewHttpClient(headers []Header) http.Client {
	return http.Client{
		Timeout: dtrack.DefaultTimeout,
		Transport: &transport{
			inner:   http.DefaultTransport,
			headers: headers,
		},
	}
}
