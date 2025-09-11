package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/n-korel/social-api/internal/auth"
	"github.com/n-korel/social-api/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	store         store.Storage
	authenticator auth.Authenticator
	config        AuthServiceConfig
}

type AuthServiceConfig struct {
	TokenExpiration time.Duration
	TokenHost       string
}

type AuthServiceInterface interface {
	CreateToken(ctx context.Context, email, password string) (string, error)
	ValidateToken(token string) (int64, error)
}

func NewAuthService(store store.Storage, authenticator auth.Authenticator, config AuthServiceConfig) *AuthService {
	return &AuthService{
		store:         store,
		authenticator: authenticator,
		config:        config,
	}
}

func (s *AuthService) CreateToken(ctx context.Context, email, password string) (string, error) {
	// Fetch User (check if user exist) from payload
	user, err := s.store.Users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := user.Password.Compare(password); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.config.TokenExpiration).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": s.config.TokenHost,
		"aud": s.config.TokenHost,
	}

	// Generate token -> add claims
	token, err := s.authenticator.GenerateToken(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (s *AuthService) ValidateToken(token string) (int64, error) {
	jwtToken, err := s.authenticator.ValidateToken(token)
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, ErrInvalidToken
	}

	return int64(userID), nil
}
