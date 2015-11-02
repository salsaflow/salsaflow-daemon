package http

import (
	// Stdlib
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
)

const (
	StatusUnprocessableEntity     = 422
	StatusUnprocessableEntityText = "Unprocessable Entity"
)

func Status(rw http.ResponseWriter, status int) {
	switch status {
	case StatusUnprocessableEntity:
		http.Error(rw, StatusUnprocessableEntityText, StatusUnprocessableEntity)
	default:
		http.Error(rw, http.StatusText(status), status)
	}
}

func Error(rw http.ResponseWriter, r *http.Request, err error) {
	Status(rw, http.StatusInternalServerError)

	// We want to mention the function that called Error,
	// hence we have to increase the number of skipped callers.
	log.NewLogger().IncreaseSkippedCallers().Error(r, err)
}
