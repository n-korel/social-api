package main

import (
	"net/http"
	"testing"

	"github.com/n-korel/social-api/internal/service"
	"github.com/n-korel/social-api/internal/store"
	"github.com/stretchr/testify/mock"
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
		mockAuthService := app.services.Auth.(*service.MockAuthService)
		mockUserService := app.services.Users.(*service.MockUserService)
		
		mockAuthService.On("ValidateToken", testToken).Return(int64(1), nil).Once()
		
		expectedUser := &store.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Role: store.Role{
				Name:  "user",
				Level: 1,
			},
		}
		mockUserService.On("GetUserByID", mock.Anything, int64(1), true).Return(expectedUser, nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		w := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, w.Code)

		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})
}