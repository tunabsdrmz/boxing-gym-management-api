package utils

import (
	"log"
	"net/http"
)

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func ForbiddenResponse(w http.ResponseWriter, r *http.Request) {
	log.Println("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	WriteJSONError(w, http.StatusForbidden, "forbidden")
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusConflict, err.Error())
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	msg := err.Error()
	if msg == "" {
		msg = "not found"
	}
	WriteJSONError(w, http.StatusNotFound, msg)
}

func UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	log.Println("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	WriteJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}