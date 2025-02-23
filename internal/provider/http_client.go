package provider

import (
	"bytes"
	"encoding/json"
	dtrack "github.com/DependencyTrack/client-go"
	"io"
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
	// Patch bugs in SDK
	if req.URL.Path == "/api/v1/repository" && (req.Method == "PUT" || req.Method == "POST") {
		// Missing authenticationRequired field when creating / updating a repository, resulting in 500 InternalServerError from API
		var repo dtrack.Repository
		json.NewDecoder(req.Body)
		err := json.NewDecoder(req.Body).Decode(&repo)
		if err != nil {
			return nil, err
		}
		type PatchedRepository struct {
			dtrack.Repository
			AuthenticationRequired bool `json:"authenticationRequired"`
		}
		patched := PatchedRepository{
			repo,
			repo.Username != "" || repo.Password != "",
		}
		bodyBuf := new(bytes.Buffer)
		err = json.NewEncoder(bodyBuf).Encode(patched)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bodyBuf)
	}
	// End patching
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
