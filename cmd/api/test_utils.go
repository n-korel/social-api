package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/n-korel/social-api/internal/auth"
	"github.com/n-korel/social-api/internal/store"
	"github.com/n-korel/social-api/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCachestore := cache.NewMockStore()

	testAuth := &auth.TestAuthenticator{}

	return &application{
		logger: logger,
		store: mockStore,
		cacheStorage: mockCachestore,
		authenticator: testAuth,
	}
}


func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	return w
}


func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}