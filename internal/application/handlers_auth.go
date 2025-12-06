package application

import (
	"net/http"

	"github.com/cdriehuys/secret-santa/internal/models"
)

func (a *Application) registerGet(w http.ResponseWriter, r *http.Request) {
	var data TemplateData = a.templateData(r)
	a.render(w, r, "register.html", data)
}

func (a *Application) registerPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	newUser := models.NewUser{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	if err := a.Users.Register(r.Context(), newUser); err != nil {
		a.serverError(w, r, "Failed to register user.", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
