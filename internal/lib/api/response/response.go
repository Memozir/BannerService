package response

import "net/http"

const (
	IncorrectDataMsg = "incorrect data"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func Error(msg string) ErrorResponse {
	return ErrorResponse{
		Error: msg,
	}
}

func SetOk(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusOK)
}

func SetDeleted(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusNonAuthoritativeInfo)
}

func SetCreated(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusCreated)
}

func SetNotFound(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusNotFound)
}

func SetBadRequest(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusBadRequest)
}

func SetUnauthorized(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusUnauthorized)
}

func SetForbidden(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusForbidden)
}

func SetInternalServerError(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusInternalServerError)
}
