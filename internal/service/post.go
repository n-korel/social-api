package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/n-korel/social-api/internal/store"
)

var (
	ErrPostNotFound = errors.New("Post not found")
)

type PostUpdateRequest struct {
	Title   *string
	Content *string
}

type PostService struct {
	store store.Storage
}

func NewPostService(store store.Storage) *PostService {
	return &PostService{
		store: store,
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID int64, title, content string, tags []string) (*store.Post, error) {
	post := &store.Post{
		Title:   title,
		Content: content,
		Tags:    tags,
		UserID:  userID,
	}

	if err := s.store.Posts.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("Failed to create post: %w", err)
	}

	return post, nil
}

func (s *PostService) GetPostByID(ctx context.Context, postID int64) (*store.Post, error) {
	post, err := s.store.Posts.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("Failed to get post: %w", err)
	}

	// Comments
	comments, err := s.store.Comments.GetByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get comments: %w", err)
	}
	post.Comments = comments

	return post, nil
}

func (s *PostService) UpdatePost(ctx context.Context, postID int64, updates PostUpdateRequest) (*store.Post, error) {
	post, err := s.store.Posts.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("Failed to get post: %w", err)
	}

	if updates.Title != nil {
		post.Title = *updates.Title
	}
	if updates.Content != nil {
		post.Content = *updates.Content
	}

	if err := s.store.Posts.Update(ctx, post); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("Failed to update post: %w", err)
	}

	return post, nil
}

func (s *PostService) DeletePost(ctx context.Context, postID int64) error {
	if err := s.store.Posts.Delete(ctx, postID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrPostNotFound
		}
		return fmt.Errorf("Failed to delete post: %w", err)
	}
	return nil
}

// Check if user can modify a post
func (s *PostService) CanUserModifyPost(ctx context.Context, user *store.User, post *store.Post, requiredRole string) (bool, error) {
	// Is the user the owner of this post?
	if post.UserID == user.ID {
		return true, nil
	}

	// Role check
	role, err := s.store.Roles.GetByName(ctx, requiredRole)
	if err != nil {
		return false, fmt.Errorf("Failed to get role: %w", err)
	}

	return user.Role.Level >= role.Level, nil
}

func (s *PostService) GetUserFeed(ctx context.Context, userID int64, query store.PaginatedFeedQuery) ([]store.PostWithMetadata, error) {
	feed, err := s.store.Posts.GetUserFeed(ctx, userID, query)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user feed: %w", err)
	}
	return feed, nil
}
