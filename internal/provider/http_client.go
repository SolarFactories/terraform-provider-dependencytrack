package provider

import (
	"bytes"
	"encoding/json"
	dtrack "github.com/DependencyTrack/client-go"
	"io"
	"net/http"
	"regexp"
)

type Header struct {
	Name  string
	Value string
}

type transport struct {
	inner   http.RoundTripper
	headers []Header
}

var (
	projectPropertyUrlRegex *regexp.Regexp = regexp.MustCompile("^/api/v1/project/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/property$")
)

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, header := range t.headers {
		req.Header.Add(header.Name, header.Value)
	}
	// Patch bugs in SDK
	if req.URL.Path == "/api/v1/repository" && (req.Method == "PUT" || req.Method == "POST") {
		// Missing authenticationRequired field when creating / updating a repository, resulting in 500 InternalServerError from API
		var repo dtrack.Repository
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
	if projectPropertyUrlRegex.MatchString(req.URL.Path) && req.Method == "DELETE" {
		// Missing PropertyType accepted by SDK method when deleting a ProjectProperty Config value
		var property dtrack.ProjectProperty
		err := json.NewDecoder(req.Body).Decode(&property)
		if err != nil {
			return nil, err
		}
		// Deleting the project property by Group and Name, so the type does not matter
		// It just needs to be able to be deserialised by the API
		property.Type = "STRING"
		bodyBuf := new(bytes.Buffer)
		err = json.NewEncoder(bodyBuf).Encode(property)
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
