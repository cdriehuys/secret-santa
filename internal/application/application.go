package application

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/secret-santa/internal/pairings"
)

const MaxNames = 100
const MaxExclusions = 3

type GiftRestrictions map[string][]string

type pairingGenerator func(GiftRestrictions) ([]pairings.Pairing, error)
type templateEngine interface {
	Render(io.Writer, string, any) error
}

type Application struct {
	Logger *slog.Logger

	PairingGenerator pairingGenerator
	Templates        templateEngine
}

func (s *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /pairings", s.pairingsGet)
	mux.HandleFunc("POST /pairings", s.pairingsPost)

	return mux
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, page string) {
	if err := a.Templates.Render(w, page, nil); err != nil {
		a.Logger.Error("Failed to render page.", "page", page, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *Application) pairingsGet(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "pairings.html")
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
