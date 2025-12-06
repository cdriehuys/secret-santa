package testutils

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/templating"
	"github.com/cdriehuys/secret-santa/ui"
)

type TestAppOpt func(a *application.Application)

type CapturingTemplateData struct {
	Data    application.TemplateData
	RawData any
}

func (d *CapturingTemplateData) Render(w io.Writer, page string, data any) error {
	d.RawData = data

	if templateData, ok := data.(application.TemplateData); ok {
		d.Data = templateData
	}

	fmt.Fprintf(w, "Render for %q captured by %#v", page, d)

	return nil
}

func CaptureTemplateData() (*CapturingTemplateData, TestAppOpt) {
	engine := &CapturingTemplateData{}

	opt := func(a *application.Application) {
		a.Templates = engine
	}

	return engine, opt
}

func WithTemplateEngine(templates application.TemplateEngine) TestAppOpt {
	return func(a *application.Application) {
		a.Templates = templates
	}
}

func NewTestApplication(t *testing.T, opts ...TestAppOpt) *application.Application {
	app := &application.Application{}

	// Apply options
	for _, opt := range opts {
		opt(app)
	}

	if app.Logger == nil {
		app.Logger = slog.New(slog.DiscardHandler)
	}

	if app.Templates == nil {
		// Default to using the embedded file system like production.
		templateFS, err := fs.Sub(ui.FS, "templates")
		if err != nil {
			t.Fatalf("failed to load templates from file system: %v", err)
		}

		templates, err := templating.NewTemplateCache(app.Logger, templateFS)
		if err != nil {
			t.Fatalf("failed to construct template cache: %v", err)
		}

		app.Templates = templates
	}

	return app
}

type TestServer struct {
	*httptest.Server
}

func NewTestServer(t *testing.T, h http.Handler) *TestServer {
	ts := httptest.NewServer(h)

	// Prevent the test server client from following redirects to allow for testing against the
	// redirect response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

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
