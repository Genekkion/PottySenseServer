package internal

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

// /htmx/settings
func (server *Server) htmxSettingsHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.htmxSettingsPanel(writer, request)
	case http.MethodPut:
		server.htmxSettingsDetailsSave(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/settings "GET"
func (server *Server) htmxSettingsPanel(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	err := server.db.QueryRow(
		`SELECT first_name, last_name
        FROM TOfficers
		WHERE ID = $1
		`, to.Id).Scan(
		&to.FirstName,
		&to.LastName,
	)

	// TODO: Change settings form
	// to only get change of tele
	// DO NOT STORE TELE HANDLE

	if err != nil {
		log.Println("htmxSettingsPanel(), db query")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/settings.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"account":        to,
	})
}

// /htmx/settings "PUT"
func (server *Server) htmxSettingsDetailsSave(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")

	if firstName != "" || lastName != "" {
		_, err := server.db.Exec(
			`UPDATE TOfficers SET
				first_name = $1,
				last_name = $2
			WHERE id = $3
			`, firstName, lastName, to.Id)

		if err != nil {
			log.Println("htmxSettingsDetailsSave(), db update")
			log.Println(err)
			genericInternalServerErrorReply(writer)
			return
		}
		log.Println("updated names")
	}

	if request.FormValue("telegram") != "" {
		err := server.redisStorage.Set(
			request.Context(),
			request.FormValue("telegram"),
			to.Id,
			time.Minute*10,
		).Err()

		if err != nil {
			log.Println("htmxSettingsDetailsSave() - set redis")
			log.Println(err)
			genericInternalServerErrorReply(writer)
			return
		}
		log.Println("updated tele")
	}

	// TODO: Change to actual template or something else
	tmpl, _ := template.New("settingsResult").Parse(
		`
        {{ if eq . "ok" }}
        <p id="settings-fadeout">Details successfully updated!</p>
        {{ else }}
        <p id="settings-fadeout">Something went wrong, please try again.</p>
        {{ end }}
    `)

	tmpl.Execute(writer, "ok")
}

// /htmx/settings/password
func (server *Server) htmxSettingsPasswordHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.htmxSettingsPasswordForm(writer, request)
	case http.MethodPut:
		server.htmxSettingsPasswordChange(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}

}

// /htmx/settings/password "GET"
func (server *Server) htmxSettingsPasswordForm(writer http.ResponseWriter,
	request *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/htmx/settingsPassword.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
	})
}

// /htmx/settings/password "PUT"
func (server *Server) htmxSettingsPasswordChange(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	oldPassword := request.FormValue("oldPassword")
	newPassword := utils.SaltPassword(request.FormValue("newPassword"))

	var passwordHash string
	err := server.db.QueryRow(
		`SELECT password
		FROM TOfficers
		WHERE id = $1
		`, to.Id).Scan(
		&passwordHash,
	)
	if err != nil {
		log.Println("htmxSettingsPasswordChange(), db query")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(utils.SaltPassword(oldPassword)),
	)
	tmpl := template.Must(template.ParseFiles("./templates/htmx/settingsPassword.html"))

	// err != nil if old password is not the same
	if err != nil {
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"status":         "old",
		})
		return
	}

	newPasswordHash, _ := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)

	_, err = server.db.Exec(
		`UPDATE TOfficers SET
        password = $1
        WHERE id = $2`,
		newPasswordHash, to.Id)

	status := "ok"
	if err != nil {

		status = "error"
	}

	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"status":         status,
	})
}
