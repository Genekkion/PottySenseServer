package internal

import (
	"html/template"
	"net/http"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/gorilla/csrf"
)

type TabListEntry struct {
	// html id
	Id          string
	Title       string
	HtmxPath    string
	RedirectUrl string
}

var (
	tabListEntries = []TabListEntry{
		{
			Id:          "tab-track",
			Title:       "Track",
			HtmxPath:    "/htmx/track",
			RedirectUrl: "/track",
		},
		{
			Id:          "tab-clients",
			Title:       "Clients",
			HtmxPath:    "/htmx/clients",
			RedirectUrl: "/clients",
		},
		{
			Id:          "tab-accounts",
			Title:       "Accounts",
			HtmxPath:    "/htmx/accounts",
			RedirectUrl: "/accounts",
		},
		{
			Id:          "tab-settings",
			Title:       "Settings",
			HtmxPath:    "/htmx/settings",
			RedirectUrl: "/settings",
		},
	}
)

// Handles templating for dashboard routes
// NOTE: Only for internal use
func (server *Server) dashboardHandler(writer http.ResponseWriter,
	request *http.Request, tabListEntry TabListEntry) {
	tmpl := template.Must(
		template.ParseFiles(
			globals.BASE_TEMPLATE,
			"./templates/htmx/dashboard.html",
		))
	tmpl.ExecuteTemplate(writer, "base", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"to":             server.getTOFromCookie(request),
		"tabListEntries": tabListEntries,
		"htmxPath":       tabListEntry.HtmxPath,
		"redirectUrl":    tabListEntry.RedirectUrl,
	})
}

// /track
func (server *Server) dashboardTrack(writer http.ResponseWriter,
	request *http.Request) {
	server.dashboardHandler(writer, request, TabListEntry{
		Id:          "tab-track",
		Title:       "Track",
		HtmxPath:    "/htmx/track",
		RedirectUrl: "/track",
	})
}

// /clients
func (server *Server) dashboardClients(writer http.ResponseWriter,
	request *http.Request) {
	server.dashboardHandler(writer, request, TabListEntry{

		Id:          "tab-clients",
		Title:       "Clients",
		HtmxPath:    "/htmx/clients",
		RedirectUrl: "/clients",
	})
}

// /accounts
// Only admins can see this page
func (server *Server) dashboardAccounts(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)
	if to.UserType != "admin" {
		writer.Header().Set("HX-Redirect",
			globals.DEFAULT_DASHBOARD_ROUTE)
		return
	}

	server.dashboardHandler(writer, request, TabListEntry{

		Id:          "tab-accounts",
		Title:       "Accounts",
		HtmxPath:    "/htmx/accounts",
		RedirectUrl: "/accounts",
	})
}

// /settings
func (server *Server) dashboardSettings(writer http.ResponseWriter,
	request *http.Request) {
	server.dashboardHandler(writer, request, TabListEntry{
		Id:          "tab-settings",
		Title:       "Settings",
		HtmxPath:    "/htmx/settings",
		RedirectUrl: "/settings",
	})
}
