package internal

import (
	"html/template"
	"net/http"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/gorilla/csrf"
)

// /dashboard
func (server *Server) dashboardHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		tmpl := template.Must(template.ParseFiles(globals.BASE_TEMPLATE))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"hxGet":          "/htmx/dashboard",
			"hxReplaceUrl":   "/dashboard",
		})

	default:
		genericMethodNotAllowedReply(writer)
	}

}

// /htmx/dashboard "GET"
func (server *Server) htmxDashboardPanel(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		tmpl := template.Must(template.ParseFiles("./templates/htmx/dashboard.html"))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"to":             server.getTOFromCookie(request),
		})
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/login
func (server *Server) htmxLoginPanel(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		tmpl := template.Must(template.ParseFiles("./templates/htmx/login.html"))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
		})
	default:
		genericMethodNotAllowedReply(writer)
	}
}
