package testutils

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/application"
)

func NewTestApplication(t *testing.T) *application.Application {
	return &application.Application{
		Logger: slog.New(slog.DiscardHandler),
	}
}

type TestServer struct {
	*httptest.Server
}

func NewTestServer(t *testing.T, h http.Handler) *TestServer {
	ts := httptest.NewServer(h)

	return &TestServer{ts}
}

type TestResponse struct {
	Status  int
	Headers http.Header
	Cookies []*http.Cookie
	Body    string
}

func MakeTestResponse(t *testing.T, res *http.Response) TestResponse {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return TestResponse{
		Status:  res.StatusCode,
		Headers: res.Header,
		Cookies: res.Cookies(),
		Body:    string(bytes.TrimSpace(body)),
	}
}

func (ts *TestServer) Get(t *testing.T, path string) TestResponse {
	req := ts.makeRequest(t, http.MethodGet, path, nil)

	return ts.doRequest(t, req)
}

func (ts *TestServer) PostForm(t *testing.T, path string, form url.Values) TestResponse {
	req := ts.makeRequest(t, http.MethodPost, path, strings.NewReader(form.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return ts.doRequest(t, req)
}

func (ts *TestServer) makeRequest(t *testing.T, method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatalf("failed to create %s request to %q: %v", method, path, err)
	}

	return req
}

func (ts *TestServer) doRequest(t *testing.T, req *http.Request) TestResponse {
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("failed to send %s request to %q: %v", req.Method, req.URL.Path, err)
	}

	defer res.Body.Close()

	return MakeTestResponse(t, res)
}
