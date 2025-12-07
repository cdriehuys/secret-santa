package application

import (
	"net/http"

	"github.com/justinas/alice"
)

func (a *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /pairings", a.pairingsGet)
	mux.HandleFunc("POST /pairings", a.pairingsPost)

	// Middleware applied to dynamic requests, ie requests that depend on the user who sent them.
	dynamic := alice.New(a.preventCSRF)

	mux.Handle("GET /{$}", dynamic.ThenFunc(a.homeGet))
	mux.Handle("GET /register", dynamic.ThenFunc(a.registerGet))
	mux.Handle("POST /register", dynamic.ThenFunc(a.registerPost))
	mux.Handle("GET /register/success", dynamic.ThenFunc(a.registerSuccess))

	// Middleware applied to all requests.
	standard := alice.New(a.RecoverPanic)

	return standard.Then(mux)
}
