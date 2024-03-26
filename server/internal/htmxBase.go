package internal

import (
	"github.com/gorilla/csrf"
	"html/template"
	"net/http"
)

func (server *Server) htmxDashboard(writer http.ResponseWriter,
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
	tmpl := template.Must(template.ParseFiles("./templates/htmx/dashboard.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"to":             to,
	})
}

func (server *Server) htmxLogin(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/login.html"))
	err := tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
	})
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
	}
}
