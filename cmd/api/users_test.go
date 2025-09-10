package main

import (
	"net/http"
	"testing"
)


func TestGetUser(t *testing.T) {
		withRedis := config{
		redisCfg: redisConfig{
			enabled: true,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Not allow unauthenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		w := executeRequest(req, mux)

		checkResponseCode(t, http.StatusUnauthorized, w.Code)
	})


	t.Run("Allow authenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer " + testToken)

		w := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, w.Code)

	})
}