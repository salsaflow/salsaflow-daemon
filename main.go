package main

import (
	// Stdlib
	"net/http"

	// Internal
	"github.com/tchap/salsaflow-daemon/internal/github"

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
	mux.Handle("/events/github", github.NewHandler(githubOptions...)

	// Set up Negroni and start listening.
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + os.Getenv("PORT"))
}
