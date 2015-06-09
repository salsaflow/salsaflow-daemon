package main

import (
	// Stdlib
	"net/http"
	"os"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/handlers/github"
	"github.com/salsaflow/salsaflow-daemon/internal/handlers/jira"

	// Vendor
	"github.com/codegangsta/negroni"
)

func main() {
	// Create the top-level mux.
	mux := http.NewServeMux()

	// Register GitHub handlers.
	var githubOptions []github.OptionFunc
	if secret := os.Getenv("GITHUB_SECRET"); secret != "" {
		githubOptions = append(githubOptions, github.SetSecret(secret))
	}
	mux.Handle("/events/github", github.NewHandler(githubOptions...))

	// Register JIRA testing handler.
	mux.Handle("/jira/me", jira.NewMeHandler())

	// Set up Negroni and start listening.
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + os.Getenv("PORT"))
}
