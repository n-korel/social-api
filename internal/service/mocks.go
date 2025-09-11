package service

import (
	"context"

	"github.com/n-korel/social-api/internal/store"
	"github.com/stretchr/testify/mock"
)

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) CreateToken(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (int64, error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.Error(1)
}

// Mock UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(ctx context.Context, username, email, password string) (*store.User, string, error) {
	args := m.Called(ctx, username, email, password)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*store.User), args.String(1), args.Error(2)
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID int64, useCache bool) (*store.User, error) {
	args := m.Called(ctx, userID, useCache)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserService) ActivateUser(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserService) FollowUser(ctx context.Context, followerID, followedID int64) error {
	args := m.Called(ctx, followerID, followedID)
	return args.Error(0)
}

func (m *MockUserService) UnfollowUser(ctx context.Context, followerID, followedID int64) error {
	args := m.Called(ctx, followerID, followedID)
	return args.Error(0)
}

// Mock PostService
type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) CreatePost(ctx context.Context, userID int64, title, content string, tags []string) (*store.Post, error) {
	args := m.Called(ctx, userID, title, content, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Post), args.Error(1)
}

func (m *MockPostService) GetPostByID(ctx context.Context, postID int64) (*store.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Post), args.Error(1)
}

func (m *MockPostService) UpdatePost(ctx context.Context, postID int64, updates PostUpdateRequest) (*store.Post, error) {
	args := m.Called(ctx, postID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Post), args.Error(1)
}

func (m *MockPostService) DeletePost(ctx context.Context, postID int64) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostService) CanUserModifyPost(ctx context.Context, user *store.User, post *store.Post, requiredRole string) (bool, error) {
	args := m.Called(ctx, user, post, requiredRole)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostService) GetUserFeed(ctx context.Context, userID int64, query store.PaginatedFeedQuery) ([]store.PostWithMetadata, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]store.PostWithMetadata), args.Error(1)
}
