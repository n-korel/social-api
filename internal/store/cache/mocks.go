package cache

import (
	"context"

	"github.com/n-korel/social-api/internal/store"
)

func NewMockStore() *Storage {
	return &Storage{
		users: &MockUserStore{},
	}
}

type MockUserStore struct {
}

func (m *MockUserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	return nil, nil
}

func (m *MockUserStore) Set(ctx context.Context, user *store.User) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, userID int64) {
}
