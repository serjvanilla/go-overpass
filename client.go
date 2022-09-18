package overpass

import (
	"net/http"
	"net/url"
)

const apiEndpoint = "https://overpass-api.de/api/interpreter"

type HTTPClient interface {
	PostForm(url string, data url.Values) (*http.Response, error)
}

// A Client manages communication with the Overpass API.
type Client struct {
	apiEndpoint string
	httpClient  HTTPClient
	semaphore   chan struct{}
}

// New returns Client instance with default overpass-api.de endpoint.
func New() Client {
	return NewWithSettings(apiEndpoint, 1, http.DefaultClient)
}

// NewWithSettings returns Client with custom settings.
func NewWithSettings(
	apiEndpoint string,
	maxParallel int,
	httpClient HTTPClient,
) Client {
	c := Client{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
		semaphore:   make(chan struct{}, maxParallel),
	}
	for i := 0; i < maxParallel; i++ {
		c.semaphore <- struct{}{}
	}

	return c
}

var DefaultClient = New()
