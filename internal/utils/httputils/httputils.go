package httputils

import (
	// Stdlib
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
)

func Error(rw http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	http.Error(rw, http.StatusText(code), code)
	log.Error(r, err)
}
