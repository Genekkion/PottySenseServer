package main

import (
	"net/http"
)

func (server *Server) addStaticRoutes() {
	fs1 := http.FileServer(http.Dir("./static"))
	server.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs1))
}
