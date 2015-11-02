package endpoint

import (
	// Stdlib
	stdLog "log"
	"net/http"

	// Internal
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	module "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"

	// Vendor
	"github.com/codegangsta/negroni"
)

const SecretQueryParameter = "secret"

type Endpoint struct{}

func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

func (ep *Endpoint) ModuleId() string {
	return module.ModuleId
}

func (ep *Endpoint) NewHandler() (http.Handler, error) {
	// Create a new mux.
	mux := http.NewServeMux()

	// Handle /events
	var activityHandler http.Handler
	if secret := config.Get().WebhookSecret; secret != "" {
		n := negroni.New()
		n.Use(newSecretMiddleware(secret))
		n.UseHandlerFunc(handleActivity)
		activityHandler = n
	} else {
		activityHandler = http.HandlerFunc(handleActivity)
		stdLog.Println("WARNING: SFD_PIVOTALTRACKER_WEBHOOK_SECRET is not set")
	}
	mux.Handle("/events", activityHandler)

	// Return the mux.
	return mux, nil
}

func newSecretMiddleware(secret string) negroni.HandlerFunc {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Check the secret query parameter.
			secretParam := r.URL.Query().Get(SecretQueryParameter)

			if secretParam != secret {
				log.Warn(r, "webhook secret mismatch: expected='%v', got='%v'\n",
					secretParam, secret)
				httputil.Status(rw, http.StatusUnauthorized)
				return
			}

			// Call the next handler.
			next(rw, r)
		})
}
