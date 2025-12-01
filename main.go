package main

import (
	"errors"
	"flag"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/pairings"
	"github.com/cdriehuys/secret-santa/internal/templating"
	"github.com/cdriehuys/secret-santa/ui"
)

var (
	liveTemplatePath string
)

func main() {
	flag.StringVar(&liveTemplatePath, "live-templates", "", "load templates from this path on each request instead of using the embedded templates")
	flag.Parse()

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

	var templates application.TemplateEngine
	if liveTemplatePath != "" {
		templates = &templating.LiveLoader{Logger: logger, BaseDir: liveTemplatePath}
	} else {
		templateFS, err := fs.Sub(ui.FS, "templates")
		if err != nil {
			panic(err)
		}

		templates, err = templating.NewTemplateCache(logger, templateFS)
		if err != nil {
			panic(err)
		}
	}

	app := application.Application{
		Logger:           logger,
		PairingGenerator: pairingGenerator,
		Templates:        templates,
	}

	s := http.Server{
		Addr:    ":8080",
		Handler: app.Routes(),
	}

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server stopped.", "error", err)
	}
}
