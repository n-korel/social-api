package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("Resource not found")
	ErrConflict          = errors.New("Resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		GetByID(context.Context, int64) (*User, error)
		Create(context.Context, *sql.Tx, *User) error
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		Activate(context.Context, string) error
	}
	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostID(context.Context, int64) ([]Comment, error)
	}
	Followers interface {
		Follow(ctx context.Context, followedID, userID int64) error
		Unfollow(ctx context.Context, followedID, userID int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostStore{
			db,
		},
		Users: &UserStore{
			db,
		},
		Comments: &CommentStore{
			db,
		},
		Followers: &FollowerStore{
			db,
		},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
