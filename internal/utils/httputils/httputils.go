package httputils

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
	log.Error(r, err)
}
