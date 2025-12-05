package application_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/application/testutils"
	"github.com/cdriehuys/secret-santa/internal/models/mocks"
)

func TestApplication_registerGet(t *testing.T) {
	templateCapture, templateOpt := testutils.CaptureTemplateData()
	app := testutils.NewTestApplication(t, templateOpt)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/register")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}

	if templateCapture.Data.IsAuthenticated {
		t.Errorf("Expected user not to be authenticated.")
	}
}

func TestApplication_registerPost(t *testing.T) {
	_, appOpt := testutils.CaptureTemplateData()
	app := testutils.NewTestApplication(t, appOpt)

	userModel := &mocks.UserModel{}
	app.Users = userModel

	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "password")

	res := ts.PostForm(t, "/register", form)

	if res.Status != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, res.Status)
	}

	if got := res.Headers.Get("Location"); got != "/" {
		t.Errorf("Expected redirect to %q, not %q", "/", got)
	}

	if userModel.RegisteredUser.Email != "test@example.com" {
		t.Errorf("Expected registered user's email to be %q, not %q", "test@example.com", userModel.RegisteredUser.Email)
	}

	if got := userModel.RegisteredUser.Password; got != "password" {
		t.Errorf("Expected registered user's password to be %q, not %q", "password", got)
	}
}
