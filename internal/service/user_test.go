package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/n-korel/social-api/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) GetByID(ctx context.Context, id int64) (*store.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*store.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *store.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *store.User, token string, exp time.Duration) error {
	args := m.Called(ctx, user, token, exp)
	return args.Error(0)
}

func (m *MockUserStore) Activate(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockFollowerStore struct {
	mock.Mock
}

func (m *MockFollowerStore) Follow(ctx context.Context, followerID, userID int64) error {
	args := m.Called(ctx, followerID, userID)
	return args.Error(0)
}

func (m *MockFollowerStore) Unfollow(ctx context.Context, followerID, userID int64) error {
	args := m.Called(ctx, followerID, userID)
	return args.Error(0)
}

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	args := m.Called(templateFile, username, email, data, isSandbox)
	return args.Int(0), args.Error(1)
}

type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) Get(ctx context.Context, id int64) (*store.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *MockUserCache) Set(ctx context.Context, user *store.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserCache) Delete(ctx context.Context, id int64) {
}

type MockCacheStorage struct {
	userCache *MockUserCache
}

func (m *MockCacheStorage) Users() UserCache {
	return m.userCache
}

func NewMockCacheStorage() *MockCacheStorage {
	return &MockCacheStorage{
		userCache: new(MockUserCache),
	}
}

// Test cases
func TestUserService_RegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		email         string
		password      string
		setupMocks    func(*MockUserStore, *MockMailer)
		expectedError error
		checkResult   func(*testing.T, *store.User, string)
	}{
		{
			name:     "successful registration",
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userStore *MockUserStore, mailer *MockMailer) {
				userStore.On("CreateAndInvite", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mailer.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(200, nil)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, user *store.User, token string) {
				assert.NotNil(t, user)
				assert.Equal(t, "testuser", user.Username)
				assert.Equal(t, "test@example.com", user.Email)
				assert.NotEmpty(t, token)
			},
		},
		{
			name:     "duplicate email",
			username: "testuser",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func(userStore *MockUserStore, mailer *MockMailer) {
				userStore.On("CreateAndInvite", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(store.ErrDuplicateEmail)
			},
			expectedError: ErrEmailAlreadyExists,
		},
		{
			name:     "email sending failure with rollback",
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userStore *MockUserStore, mailer *MockMailer) {
				userStore.On("CreateAndInvite", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mailer.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(0, errors.New("mail server error"))
				userStore.On("Delete", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedError: errors.New("Failed to send activation email"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockUserStore := new(MockUserStore)
			mockMailer := new(MockMailer)
			
			mockStorage := store.Storage{
				Users: mockUserStore,
			}
			
			service := NewUserService(
				mockStorage,
				nil, // cache not needed for this test
				mockMailer,
				UserServiceConfig{
					FrontendURL:     "http://localhost:3000",
					MailExpiration:  24 * time.Hour,
					IsProductionEnv: false,
				},
			)

			// Setup mocks
			tt.setupMocks(mockUserStore, mockMailer)

			// Execute
			user, token, err := service.RegisterUser(
				context.Background(),
				tt.username,
				tt.email,
				tt.password,
			)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, user, token)
				}
			}

			// Verify mock expectations
			mockUserStore.AssertExpectations(t)
			mockMailer.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByID_WithCache(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	expectedUser := &store.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	t.Run("cache hit", func(t *testing.T) {
		// Setup
		mockUserStore := new(MockUserStore)
		mockCacheStorage := NewMockCacheStorage()
		
		mockStorage := store.Storage{
			Users: mockUserStore,
		}
		
		service := NewUserService(mockStorage, mockCacheStorage, nil, UserServiceConfig{})

		// Setup cache to return user
		mockCacheStorage.userCache.On("Get", ctx, userID).Return(expectedUser, nil)

		// Execute
		user, err := service.GetUserByID(ctx, userID, true)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		
		// Verify DB was not called
		mockUserStore.AssertNotCalled(t, "GetByID")
		mockCacheStorage.userCache.AssertExpectations(t)
	})

	t.Run("cache miss", func(t *testing.T) {
		// Setup
		mockUserStore := new(MockUserStore)
		mockCacheStorage := NewMockCacheStorage()
		
		mockStorage := store.Storage{
			Users: mockUserStore,
		}
		
		service := NewUserService(mockStorage, mockCacheStorage, nil, UserServiceConfig{})

		// Setup cache miss and DB hit
		mockCacheStorage.userCache.On("Get", ctx, userID).Return(nil, nil)
		mockUserStore.On("GetByID", ctx, userID).Return(expectedUser, nil)
		mockCacheStorage.userCache.On("Set", ctx, expectedUser).Return(nil)

		// Execute
		user, err := service.GetUserByID(ctx, userID, true)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		
		// Verify all calls
		mockUserStore.AssertExpectations(t)
		mockCacheStorage.userCache.AssertExpectations(t)
	})
}

func TestUserService_FollowUser(t *testing.T) {
	ctx := context.Background()
	followerID := int64(1)
	followedID := int64(2)
	
	follower := &store.User{ID: followerID, Username: "follower"}
	followed := &store.User{ID: followedID, Username: "followed"}

	t.Run("successful follow", func(t *testing.T) {
		// Setup
		mockUserStore := new(MockUserStore)
		mockFollowerStore := new(MockFollowerStore)
		
		mockStorage := store.Storage{
			Users:     mockUserStore,
			Followers: mockFollowerStore,
		}
		
		service := NewUserService(mockStorage, nil, nil, UserServiceConfig{})

		// Setup mocks
		mockUserStore.On("GetByID", ctx, followerID).Return(follower, nil)
		mockUserStore.On("GetByID", ctx, followedID).Return(followed, nil)
		mockFollowerStore.On("Follow", ctx, followerID, followedID).Return(nil)

		// Execute
		err := service.FollowUser(ctx, followerID, followedID)

		// Assert
		assert.NoError(t, err)
		mockUserStore.AssertExpectations(t)
		mockFollowerStore.AssertExpectations(t)
	})

	t.Run("cannot follow self", func(t *testing.T) {
		// Setup
		mockUserStore := new(MockUserStore)
		mockStorage := store.Storage{
			Users: mockUserStore,
		}
		
		service := NewUserService(mockStorage, nil, nil, UserServiceConfig{})

		// Setup mock
		mockUserStore.On("GetByID", ctx, followerID).Return(follower, nil)
		mockUserStore.On("GetByID", ctx, followerID).Return(follower, nil)

		// Execute
		err := service.FollowUser(ctx, followerID, followerID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrCannotFollowSelf, err)
		mockUserStore.AssertExpectations(t)
	})

	t.Run("already following", func(t *testing.T) {
		// Setup
		mockUserStore := new(MockUserStore)
		mockFollowerStore := new(MockFollowerStore)
		
		mockStorage := store.Storage{
			Users:     mockUserStore,
			Followers: mockFollowerStore,
		}
		
		service := NewUserService(mockStorage, nil, nil, UserServiceConfig{})

		// Setup mocks
		mockUserStore.On("GetByID", ctx, followerID).Return(follower, nil)
		mockUserStore.On("GetByID", ctx, followedID).Return(followed, nil)
		mockFollowerStore.On("Follow", ctx, followerID, followedID).Return(store.ErrConflict)

		// Execute
		err := service.FollowUser(ctx, followerID, followedID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrAlreadyFollowing, err)
		mockUserStore.AssertExpectations(t)
		mockFollowerStore.AssertExpectations(t)
	})
}