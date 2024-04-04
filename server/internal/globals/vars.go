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

	// WARN: Harcoded for single toilet with id of 1
	TOILETS_URL = map[int]string{
	
	}
)
