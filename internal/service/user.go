package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/n-korel/social-api/internal/mailer"
	"github.com/n-korel/social-api/internal/store"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrUsernameAlreadyExists  = errors.New("username already exists")
	ErrInvalidActivationToken = errors.New("invalid or expired activation token")
	ErrCannotFollowSelf       = errors.New("cannot follow yourself")
	ErrAlreadyFollowing       = errors.New("already following this user")
)

type UserService struct {
	store  store.Storage
	cache  CacheStorage
	mailer mailer.Client
	config UserServiceConfig
}

type UserServiceConfig struct {
	FrontendURL     string
	MailExpiration  time.Duration
	IsProductionEnv bool
}

type UserServiceInterface interface {
	RegisterUser(ctx context.Context, username, email, password string) (*store.User, string, error)
	GetUserByID(ctx context.Context, userID int64, useCache bool) (*store.User, error)
	ActivateUser(ctx context.Context, token string) error
	FollowUser(ctx context.Context, followerID, followedID int64) error
	UnfollowUser(ctx context.Context, followerID, followedID int64) error
}

type UserCache interface {
	Get(context.Context, int64) (*store.User, error)
	Set(context.Context, *store.User) error
	Delete(context.Context, int64)
}

type CacheStorage interface {
	Users() UserCache
}

func NewUserService(store store.Storage, cache CacheStorage, mailer mailer.Client, config UserServiceConfig) *UserService {
	return &UserService{
		store:  store,
		cache:  cache,
		mailer: mailer,
		config: config,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, username, email, password string) (*store.User, string, error) {
	user := &store.User{
		Username: username,
		Email:    email,
		Role: store.Role{
			Name: "user",
		},
	}

	// Hash user password
	if err := user.Password.Set(password); err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Hash token for storage but keep plain token for email
	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// Store user with invitation
	err := s.store.Users.CreateAndInvite(ctx, user, hashToken, s.config.MailExpiration)
	if err != nil {
		return nil, "", s.handleUserCreationError(err)
	}

	// Send email
	if err := s.sendActivationEmail(user, plainToken); err != nil {
		// Rollback if email fails
		if delErr := s.store.Users.Delete(ctx, user.ID); delErr != nil {
			return nil, "", fmt.Errorf("failed to send email and rollback: %w, rollback error: %v", err, delErr)
		}
		return nil, "", fmt.Errorf("failed to send activation email: %w", err)
	}

	return user, plainToken, nil
}

func (s *UserService) ActivateUser(ctx context.Context, token string) error {
	err := s.store.Users.Activate(ctx, token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrInvalidActivationToken
		}
		return fmt.Errorf("failed to activate user: %w", err)
	}
	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64, useCache bool) (*store.User, error) {
	if !useCache || s.cache == nil {
		return s.getUserFromDB(ctx, userID)
	}

	// Try cache
	user, err := s.cache.Users().Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("cache error: %w", err)
	}

	if user == nil {
		// Cache miss - get from DB
		user, err = s.getUserFromDB(ctx, userID)
		if err != nil {
			return nil, err
		}

		// Update cache
		if cacheErr := s.cache.Users().Set(ctx, user); cacheErr != nil {
		}
	}

	return user, nil
}

func (s *UserService) FollowUser(ctx context.Context, followerID, followedID int64) error {
	// Validate that follower user exist
	if _, err := s.getUserFromDB(ctx, followerID); err != nil {
		return fmt.Errorf("follower not found: %w", err)
	}

	// Validate that followed user exist
	if _, err := s.getUserFromDB(ctx, followedID); err != nil {
		return fmt.Errorf("followed user not found: %w", err)
	}

	// Self-following
	if followerID == followedID {
		return ErrCannotFollowSelf
	}

	err := s.store.Followers.Follow(ctx, followerID, followedID)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return ErrAlreadyFollowing
		}
		return fmt.Errorf("failed to follow user: %w", err)
	}

	return nil
}

func (s *UserService) UnfollowUser(ctx context.Context, followerID, followedID int64) error {
	err := s.store.Followers.Unfollow(ctx, followerID, followedID)
	if err != nil {
		return fmt.Errorf("failed to unfollow user: %w", err)
	}
	return nil
}

func (s *UserService) getUserFromDB(ctx context.Context, userID int64) (*store.User, error) {
	user, err := s.store.Users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) sendActivationEmail(user *store.User, token string) error {
	activationURL := fmt.Sprintf("%s/confirm/%s", s.config.FrontendURL, token)

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	_, err := s.mailer.Send(
		mailer.UserWelcomeTemplate,
		user.Username,
		user.Email,
		vars,
		!s.config.IsProductionEnv,
	)

	return err
}

func (s *UserService) handleUserCreationError(err error) error {
	switch err {
	case store.ErrDuplicateEmail:
		return ErrEmailAlreadyExists
	case store.ErrDuplicateUsername:
		return ErrUsernameAlreadyExists
	default:
		return fmt.Errorf("failed to create user: %w", err)
	}
}
