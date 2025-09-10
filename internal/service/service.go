package service

import (
	"github.com/n-korel/social-api/internal/auth"
	"github.com/n-korel/social-api/internal/mailer"
	"github.com/n-korel/social-api/internal/store"
)

type Services struct {
	Users *UserService
	Posts *PostService
	Auth  *AuthService
}

func NewServices(
	store store.Storage,
	cache CacheStorage,
	mailer mailer.Client,
	authenticator auth.Authenticator,
	userConfig UserServiceConfig,
	authConfig AuthServiceConfig,
) *Services {
	return &Services{
		Users: NewUserService(store, cache, mailer, userConfig),
		Posts: NewPostService(store),
		Auth:  NewAuthService(store, authenticator, authConfig),
	}
}
