package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read the Auth Header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("Authorization header is missing"))
				return
			}

			// Parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("Authorization header is malformed"))
				return
			}

			// Decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			// Check the Credentials
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass

			credential := strings.SplitN(string(decoded), ":", 2)
			if len(credential) != 2 || credential[0] != username || credential[1] != pass {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("Invalid Credemtials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
