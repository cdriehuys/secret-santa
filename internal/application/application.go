package application

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/secret-santa/internal/pairings"
)

const MaxNames = 100
const MaxExclusions = 3

type GiftRestrictions map[string][]string

type pairingGenerator func(GiftRestrictions) ([]pairings.Pairing, error)

type Application struct {
	Logger *slog.Logger

	PairingGenerator pairingGenerator
}

func (s *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /pairings", s.pairingsPost)

	return mux
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
