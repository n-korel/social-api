package cache

import (
	"github.com/n-korel/social-api/internal/service"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	users service.UserCache
}

func (s *Storage) Users() service.UserCache {
	return s.users
}

func NewRedisStorage(rdb *redis.Client) *Storage {
	return &Storage{
		users: &UserStore{
			rdb: rdb,
		},
	}
}
