package globals

var (
	FLAG_VERBOSE bool
	RUN          bool

	// All routes in UNPROTECTED_ROUTES will NOT
	// be CSRF protected
	UNPROTECTED_ROUTES = []string{
		"/ext/api",
		"/ext/bot",
	}
)
