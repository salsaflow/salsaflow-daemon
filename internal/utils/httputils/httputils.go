package httputils

import (
	"log"
	"net/http"
)

func Error(rw http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	http.Error(rw, http.StatusText(code), code)
	log.Println("Error:", err)
}
