package main

import (
	"errors"
	"net/http"

	"github.com/n-korel/social-api/internal/service"
)

func (app *application) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	// User service errors
	case errors.Is(err, service.ErrUserNotFound):
		app.notFoundResponse(w, r, err)
	case errors.Is(err, service.ErrEmailAlreadyExists):
		app.conflictResponse(w, r, err)
	case errors.Is(err, service.ErrUsernameAlreadyExists):
		app.conflictResponse(w, r, err)
	case errors.Is(err, service.ErrInvalidActivationToken):
		app.badRequestResponse(w, r, err)
	case errors.Is(err, service.ErrCannotFollowSelf):
		app.badRequestResponse(w, r, err)
	case errors.Is(err, service.ErrAlreadyFollowing):
		app.conflictResponse(w, r, err)

	// Auth service errors
	case errors.Is(err, service.ErrInvalidCredentials):
		app.unauthorizedErrorResponse(w, r, err)
	case errors.Is(err, service.ErrInvalidToken):
		app.unauthorizedErrorResponse(w, r, err)

	// Post service errors
	case errors.Is(err, service.ErrPostNotFound):
		app.notFoundResponse(w, r, err)

	// Default internal server error
	default:
		app.internalServerError(w, r, err)
	}
}

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "The Server encountered a problem")
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("Forbidden", "method", r.Method, "path", r.URL.Path, "error")

	writeJSONError(w, http.StatusForbidden, "Forbidden")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("Bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("Conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("Not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, "Not found")
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("Unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("Unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
}
