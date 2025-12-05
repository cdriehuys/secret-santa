package application_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/application/testutils"
	"github.com/cdriehuys/secret-santa/internal/pairings"
)

func TestApplication_homeGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "")
	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}

	if res.Body == "" {
		t.Error("Expected non-empty body.")
	}
}

func TestApplication_pairingsPost(t *testing.T) {
	names := []string{"Bob", "Jane"}
	fakeGenerator := func(restrictions application.GiftRestrictions) ([]pairings.Pairing, error) {
		pairs := []pairings.Pairing{
			{From: "Bob", To: "Jane"},
			{From: "Jane", To: "Bob"},
		}

		return pairs, nil
	}

	app := testutils.NewTestApplication(t)
	app.PairingGenerator = fakeGenerator

	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	form := url.Values{}
	for i, name := range names {
		form.Add(fmt.Sprintf("name[%d]", i), name)
	}

	res := ts.PostForm(t, "/pairings", form)

	if got := res.Status; got != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, got)
	}

	assertContains(t, res.Body, "Bob -> Jane")
	assertContains(t, res.Body, "Jane -> Bob")
}

func TestApplication_pairingsPostWithExclusions(t *testing.T) {
	people := map[string][]string{
		"Ross":     {"Joey"},
		"Joey":     {"Chandler"},
		"Chandler": {"Ross"},
	}

	fakeGenerator := func(application.GiftRestrictions) ([]pairings.Pairing, error) {
		pairs := []pairings.Pairing{
			{From: "Ross", To: "Chandler"},
			{From: "Joey", To: "Ross"},
			{From: "Chandler", To: "Joey"},
		}

		return pairs, nil
	}

	app := testutils.NewTestApplication(t)
	app.PairingGenerator = fakeGenerator

	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	form := url.Values{}
	i := 0
	for person, restrictions := range people {
		form.Add(fmt.Sprintf("name[%d]", i), person)

		for j, restriction := range restrictions {
			form.Add(fmt.Sprintf("name[%d].exclusion[%d]", i, j), restriction)
		}

		i++
	}

	res := ts.PostForm(t, "/pairings", form)

	if got := res.Status; got != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, got)
	}

	assertContains(t, res.Body, "Ross -> Chandler")
	assertContains(t, res.Body, "Joey -> Ross")
	assertContains(t, res.Body, "Chandler -> Joey")
}

func assertContains(t *testing.T, haystack string, needle string) {
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected to find %q in %q", needle, haystack)
	}
}
