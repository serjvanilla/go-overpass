package overpass

import "net/http"

const apiEndpoint = "https://overpass-api.de/api/interpreter"

// A Client manages communication with the Overpass API.
type Client struct {
	apiEndpoint string
	httpClient  *http.Client
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
	httpClient *http.Client,
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
