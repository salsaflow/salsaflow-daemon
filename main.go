package main

import (
	// Stdlib
	"log"
	"net/http"
	"os"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/endpoints"

	// Vendor
	"github.com/codegangsta/negroni"
)

func main() {
	// Register the module endpoints with the main mux.
	var nuked bool
	mux := http.NewServeMux()
	for _, endpoint := range endpoints.Endpoints() {
		handler, err := endpoint.NewHandler()
		if err != nil {
			log.Println(err)
			nuked = true
		}

		prefix := "/modules/" + endpoint.ModuleId()
		mux.Handle(prefix+"/", http.StripPrefix(prefix, handler))
	}
	if nuked {
		os.Exit(1)
	}

	// Set up Negroni and start listening.
	n := negroni.Classic()
	n.Use(newRewriteObsoletePathsMiddleware())
	n.UseHandler(mux)
	n.Run(":" + os.Getenv("PORT"))
}

func newRewriteObsoletePathsMiddleware() negroni.Handler {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Handle obsolete paths by rewriting them internally to the new paths.
			switch r.URL.Path {
			case "/events/github":
				r.URL.Path = "/modules/salsaflow.modules.codereview.github/events"
			case "/events/pivotaltracker":
				r.URL.Path = "/modules/salsaflow.modules.issuetracking.pivotaltracker/events"
			}

			// Pass the request to the next handler.
			next(rw, r)
		})
}
