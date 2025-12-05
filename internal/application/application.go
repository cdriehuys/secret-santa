package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/secret-santa/internal/models"
	"github.com/cdriehuys/secret-santa/internal/pairings"
)

const MaxNames = 100
const MaxExclusions = 3

type GiftRestrictions map[string][]string

type pairingGenerator func(GiftRestrictions) ([]pairings.Pairing, error)
type TemplateEngine interface {
	Render(io.Writer, string, any) error
}

type UserModel interface {
	Register(context.Context, models.NewUser) error
}

type TemplateData struct {
	IsAuthenticated bool
}

type Application struct {
	Logger *slog.Logger

	PairingGenerator pairingGenerator
	Templates        TemplateEngine

	Users UserModel
}

func (s *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", s.homeGet)
	mux.HandleFunc("GET /pairings", s.pairingsGet)
	mux.HandleFunc("POST /pairings", s.pairingsPost)
	mux.HandleFunc("GET /register", s.registerGet)
	mux.HandleFunc("POST /register", s.registerPost)

	return mux
}

func (a *Application) templateData(r *http.Request) TemplateData {
	return TemplateData{}
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, page string, data TemplateData) {
	if err := a.Templates.Render(w, page, data); err != nil {
		a.Logger.Error("Failed to render page.", "page", page, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *Application) homeGet(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "home.html", a.templateData(r))
}

func (a *Application) pairingsGet(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "pairings.html", a.templateData(r))
}

func (s *Application) pairingsPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	restrictions := GiftRestrictions{}

	for i := range MaxNames {
		nameKey := fmt.Sprintf("name[%d]", i)
		name := r.FormValue(nameKey)

		if name == "" {
			break
		}

		var exclusions []string
		for j := range MaxExclusions {
			exclusionKey := fmt.Sprintf("%s.exclusion[%d]", nameKey, j)
			exclusion := r.FormValue(exclusionKey)

			if exclusion == "" {
				break
			}

			exclusions = append(exclusions, exclusion)
		}

		restrictions[name] = exclusions
	}

	pairs, err := s.PairingGenerator(restrictions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Pairings:")
	for _, pair := range pairs {
		fmt.Fprintf(w, "%s -> %s\n", pair.From, pair.To)
	}
}
