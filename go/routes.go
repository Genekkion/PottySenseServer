package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

const baseTemplate = "./templates/base.html"

func (server *Server) dashboardPage(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusBadRequest,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return
	} else if !server.isValidSession(writer, request) {
        log.Println("redirecting to index")
		http.Redirect(writer, request, "/", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles(baseTemplate))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"hxGet":          "/htmx/dashboard",
		"hxReplaceUrl":   "/dashboard",
	})
}
