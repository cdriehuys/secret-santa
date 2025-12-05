package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/email"
	"github.com/cdriehuys/secret-santa/internal/models"
	"github.com/cdriehuys/secret-santa/internal/models/queries"
	"github.com/cdriehuys/secret-santa/internal/pairings"
	"github.com/cdriehuys/secret-santa/internal/security"
	"github.com/cdriehuys/secret-santa/internal/templating"
	"github.com/cdriehuys/secret-santa/ui"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	liveEmailTemplatePath string
	liveTemplatePath      string
)

func main() {
	flag.StringVar(&liveEmailTemplatePath, "live-email-templates", "", "load email templates from this path for each request instead of using the embedded templates")
	flag.StringVar(&liveTemplatePath, "live-templates", "", "load UI templates from this path for each request instead of using the embedded templates")
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

	var emailTemplates application.TemplateEngine
	if liveTemplatePath != "" {
		emailTemplates = &templating.LiveLoader{Logger: logger, BaseDir: liveEmailTemplatePath}
	} else {
		templateFS, err := fs.Sub(ui.EmailFS, "emails")
		if err != nil {
			panic(err)
		}

		emailTemplates, err = templating.NewTemplateCache(logger, templateFS)
		if err != nil {
			panic(err)
		}
	}

	var uiTemplates application.TemplateEngine
	if liveTemplatePath != "" {
		uiTemplates = &templating.LiveLoader{Logger: logger, BaseDir: liveTemplatePath}
	} else {
		templateFS, err := fs.Sub(ui.FS, "templates")
		if err != nil {
			panic(err)
		}

		uiTemplates, err = templating.NewTemplateCache(logger, templateFS)
		if err != nil {
			panic(err)
		}
	}

	emailer := email.NewConsoleMailer(os.Stdout)
	sender := "no-reply@localhost"
	baseDomain, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}

	emailVerifier := application.NewEmailVerifier(logger, emailer, emailTemplates, baseDomain, sender)

	connString := os.Getenv("DB_CONN")
	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	defer dbPool.Close()

	queries := queries.New(dbPool)

	users := models.NewUserModel(logger, emailVerifier, security.Argon2IDHasher{}, security.TokenGenerator{}, models.PoolWrapper{Pool: dbPool}, models.UserQueriesWrapper{Queries: queries})

	app := application.Application{
		Logger:           logger,
		PairingGenerator: pairingGenerator,
		Templates:        uiTemplates,

		Users: users,
	}

	s := http.Server{
		Addr:    ":8080",
		Handler: app.Routes(),
	}

	logger.Info("Starting web server.", "addr", s.Addr)

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server stopped.", "error", err)
	}
}
