package overpass

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const ApiOutputFormatJson string = "json"

type overpassResponse struct {
	OSM3S struct {
		TimestampOSMBase time.Time `json:"timestamp_osm_base"`
	} `json:"osm3s"`
	Elements []overpassJsonResponse `json:"elements"`
}

// Query send request to OverpassAPI with provided querystring.
func (c *Client) Query(query string) (Result, error) {
	body, err := c.httpPost(query)
	if err != nil {
		return Result{}, err
	}

	outF := getOutputFormatFromQuery(query)
	if outF == ApiOutputFormatJson {
		return unmarshalJson(body)
	}
	return unmarshalXml(body)
}

func (c *Client) httpPost(query string) ([]byte, error) {
	<-c.semaphore
	defer func() { c.semaphore <- struct{}{} }()

	resp, err := c.httpClient.PostForm(
		c.apiEndpoint,
		url.Values{"data": []string{query}},
	)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("overpass engine error: %w", &ServerError{resp.StatusCode, body})
	}

	return body, nil
}

// Query runs query with default client.
func Query(query string) (Result, error) {
	return DefaultClient.Query(query)
}

type ServerError struct {
	StatusCode int
	Body       []byte
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

func getOutputFormatFromQuery(query string) string {
	re := regexp.MustCompile("\\[out:([a-z]+)]")
	match := re.FindStringSubmatch(query)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}
