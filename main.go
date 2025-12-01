package main

import (
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/pairings"
	"github.com/cdriehuys/secret-santa/internal/templating"
)

func main() {
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	pairingGenerator := func(restrictions application.GiftRestrictions) ([]pairings.Pairing, error) {
		graph := pairings.NewGraphFromExclusions(restrictions)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		return graph.Pairings(r)
	}

	templates := templating.LiveLoader{Logger: logger, BaseDir: "./ui/templates"}

	app := application.Application{
		Logger:           logger,
		PairingGenerator: pairingGenerator,
		Templates:        &templates,
	}

	s := http.Server{
		Addr:    ":8080",
		Handler: app.Routes(),
	}

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server stopped.", "error", err)
	}
}
