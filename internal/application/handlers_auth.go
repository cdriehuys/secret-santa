package application

import (
	"net/http"

	"github.com/cdriehuys/secret-santa/internal/models"
)

func (a *Application) registerGet(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "register.html", a.templateData(r))
}

func (a *Application) registerPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	newUser := models.NewUser{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}
	a.Users.Register(r.Context(), newUser)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
