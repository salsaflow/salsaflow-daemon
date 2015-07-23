package github

const (
	statusUnprocessableEntity     = 422
	statusUnprocessableEntityText = "Unprocessable Entity"
)

func httpStatus(rw http.ResponseWriter, status int) {
	switch status {
	case statusUnprocessableEntity:
		http.Error(rw, statusUnprocessableEntityText, statusUnprocessableEntity)
	default:
		http.Error(rw, http.StatusText(status), status)
	}
}
