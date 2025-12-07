package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/secret-santa/internal/models"
	"github.com/cdriehuys/secret-santa/internal/pairings"
	"github.com/justinas/nosurf"
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
	CSRFToken       string
}

type Application struct {
	Logger *slog.Logger

	PairingGenerator pairingGenerator
	Templates        TemplateEngine

	Users UserModel
}

func (a *Application) templateData(r *http.Request) TemplateData {
	return TemplateData{
		CSRFToken: nosurf.Token(r),
	}
}

func (a *Application) serverError(w http.ResponseWriter, r *http.Request, message string, err error, attrs ...any) {
	attrs = append(attrs, "error", err)
	a.Logger.ErrorContext(r.Context(), message, attrs...)

	w.WriteHeader(http.StatusInternalServerError)
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, page string, data TemplateData) {
	if err := a.Templates.Render(w, page, data); err != nil {
		a.serverError(w, r, "Failed to render page.", err, "page", page)
	}
}

func (a *Application) homeGet(w http.ResponseWriter, r *http.Request) {
	var data TemplateData = a.templateData(r)
	a.render(w, r, "home.html", data)
}

func (a *Application) pairingsGet(w http.ResponseWriter, r *http.Request) {
	var data TemplateData = a.templateData(r)
	a.render(w, r, "pairings.html", data)
}

func (a *Application) pairingsPost(w http.ResponseWriter, r *http.Request) {
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

	pairs, err := a.PairingGenerator(restrictions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Pairings:")
	for _, pair := range pairs {
		fmt.Fprintf(w, "%s -> %s\n", pair.From, pair.To)
	}
}
