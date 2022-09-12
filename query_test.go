package overpass

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"testing/iotest"
)

func TestQueryErrors(t *testing.T) {
	testCases := []struct {
		res  *http.Response
		err  error
		want string
	}{
		{nil, errors.New("request fail"), "http error: request fail"},
		{&http.Response{StatusCode: 400, Body: io.NopCloser(bytes.NewReader(nil))}, nil, "overpass engine error: 400 Bad Request"},
		{&http.Response{StatusCode: 200, Body: io.NopCloser(iotest.ErrReader(errors.New("read fail")))}, nil, "http error: read fail"},
		{&http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}, nil, "overpass engine error: unexpected end of JSON input"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			cli := NewWithSettings(apiEndpoint, 1, &mockHttpClient{tc.res, tc.err})
			if _, err := cli.Query(""); err == nil {
				t.Fatal("unexpected success")
			} else if err.Error() != tc.want {
				t.Fatalf("%s != %s", err.Error(), tc.want)
			} else if err = errors.Unwrap(err); err == nil {
				t.Fatal("expected wrapped error")
			}
		})
	}
}

type mockHttpClient struct {
	res *http.Response
	err error
}

func (m *mockHttpClient) PostForm(string, url.Values) (res *http.Response, err error) {
	return m.res, m.err
}
