package internal

import (
	"html/template"
	"net/http"

	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

// /htmx/settings handler
func (server *Server) htmxSettingsHandler(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {

	case http.MethodGet:
		server.htmxSettings(writer, request)

	case http.MethodPost:

	default:
		http.Error(writer, "Method not allowed.",
			http.StatusMethodNotAllowed)
	}
}

func (server *Server) htmxSettings(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	to, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}
	db := server.dbStorage

	row := db.QueryRow(
		`SELECT first_name, last_name, telegram
        FROM TOfficers WHERE ID = $1`, to.Id)
	var firstName string
	var lastName string
	var telegram string
	err = row.Scan(&firstName, &lastName, &telegram)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/settings.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"account": TO{
			FirstName: firstName,
			LastName:  lastName,
			Telegram:  telegram,
		},
	})
}

func (server *Server) htmxSettingsDetailsSave(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "POST" {
		http.Error(
			writer,
			"Invalid request method",
			http.StatusBadRequest,
		)
		return
	}

	to, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")
	telegram := request.FormValue("telegram")

	db := server.dbStorage
	_, err = db.Exec(
		`UPDATE TOfficers SET
        first_name = $1,
        last_name = $2,
        telegram = $3
        WHERE id = $4`,
		firstName, lastName, telegram, to.Id)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}
	tmpl, err := template.New("settingsResult").Parse(
		`
        {{ if eq . "ok" }}
        <p id="settings-fadeout">Details successfully updated!</p>
        {{ else }}
        <p id="settings-fadeout">Something went wrong, please try again.</p>
        {{ end }}
    `)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	tmpl.Execute(writer, "ok")
}

func (server *Server) htmxSettingsPasswordForm(writer http.ResponseWriter,
	request *http.Request) {
	_, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/settingsPassword.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"status":         "nil",
	})

}

func (server *Server) htmxSettingsPasswordHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case "GET":
		server.htmxSettingsPasswordForm(writer, request)
		return
	case "POST":
		server.htmxSettingsPasswordChange(writer, request)
		return
	default:
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
	}

}
func (server *Server) htmxSettingsPasswordChange(writer http.ResponseWriter,
	request *http.Request) {
	to, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	oldPassword := request.FormValue("oldPassword")
	newPassword := utils.SaltPassword(request.FormValue("newPassword"))

	db := server.dbStorage

	var passwordHash string
	row := db.QueryRow("SELECT password FROM TOfficers WHERE id = $1", to.Id)
	err = row.Scan(&passwordHash)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(utils.SaltPassword(oldPassword)),
	)
	tmpl := template.Must(template.ParseFiles("./templates/htmx/settingsPassword.html"))
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

	_, err = db.Exec(
		`UPDATE TOfficers SET
        password = $1
        WHERE id = $2`,
		newPasswordHash, to.Id)
	if err != nil {
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"status":         "error",
		})
		return
	}
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"status":         "ok",
	})
}
