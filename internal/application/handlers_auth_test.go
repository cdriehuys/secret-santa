package application_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/application/testutils"
	"github.com/cdriehuys/secret-santa/internal/models"
	"github.com/cdriehuys/secret-santa/internal/models/mocks"
)

func TestApplication_registerGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/register")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}
}

type CapturingTemplateEngine[T any] struct {
	RenderError error

	RenderedName    string
	RenderedData    T
	RenderedRawData any
}

func (e *CapturingTemplateEngine[T]) Render(w io.Writer, name string, data any) error {
	e.RenderedName = name
	e.RenderedRawData = data

	if structuredData, ok := data.(T); ok {
		e.RenderedData = structuredData
	}

	// Have to return error before writing, otherwise the response will automatically get a 200
	// response.
	if e.RenderError != nil {
		return e.RenderError
	}

	fmt.Fprintf(w, "%#v", data)

	return nil
}

type WantRedirect struct {
	Status   int
	Location string
}

func TestApplication_registerPost(t *testing.T) {
	defaultEmail := "test@example.com"
	defaultPassword := "tops3cret"

	testCases := []struct {
		name           string
		templates      CapturingTemplateEngine[application.TemplateData]
		users          mocks.UserModel
		email          string
		password       string
		wantStatus     int
		wantRegistered models.NewUser
		wantRedirect   *WantRedirect
	}{
		{
			name:           "successful registration",
			email:          defaultEmail,
			password:       defaultPassword,
			wantRegistered: models.NewUser{Email: defaultEmail, Password: defaultPassword},
			wantRedirect:   &WantRedirect{Status: http.StatusSeeOther, Location: "/register/success"},
		},
		{
			name: "registration server error",
			users: mocks.UserModel{
				RegisterError: errors.New("registration failed"),
			},
			email:          defaultEmail,
			password:       defaultPassword,
			wantRegistered: models.NewUser{Email: defaultEmail, Password: defaultPassword},
			wantStatus:     http.StatusInternalServerError,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)

			app.Templates = &tt.templates
			app.Users = &tt.users

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			form := csrfFormValues(t, app, ts, "/register")
			form.Add("email", tt.email)
			form.Add("password", tt.password)

			res := ts.PostForm(t, "/register", form)

			if tt.wantStatus != 0 && res.Status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.Status)
			}

			if got := tt.users.RegisteredUser.Email; got != tt.wantRegistered.Email {
				t.Errorf("Expected registered user email %q, got %q", tt.wantRegistered.Email, got)
			}

			if got := tt.users.RegisteredUser.Password; got != tt.wantRegistered.Password {
				t.Errorf("Expected registered user password %q, got %q", tt.wantRegistered.Password, got)
			}

			if want := tt.wantRedirect; want != nil {
				if res.Status != want.Status {
					t.Errorf("Expected status %d, got %d", want.Status, res.Status)
				}

				if got := res.Headers.Get("Location"); got != want.Location {
					t.Errorf("Expected redirect location %q, got %q", want.Location, got)
				}
			}
		})
	}
}

func TestApplication_registerSuccess(t *testing.T) {
	testCases := []struct {
		name       string
		templates  CapturingTemplateEngine[application.TemplateData]
		wantStatus int
	}{
		{
			name:       "successful render",
			wantStatus: http.StatusOK,
		},
		{
			name: "render error",
			templates: CapturingTemplateEngine[application.TemplateData]{
				RenderError: errors.New("rendering failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)
			app.Templates = &tt.templates

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			res := ts.Get(t, "/register/success")

			if res.Status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.Status)
			}
		})
	}
}
