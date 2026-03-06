package interfaces

import (
    "context"
    "task-manager/internal/domain"
)

// TaskRepository defines all task-related database operations
type TaskRepository interface {
    Create(ctx context.Context, task *domain.Task) error
    GetByID(ctx context.Context, id string) (*domain.Task, error)
    GetByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Task, int64, error)
    GetByStatus(ctx context.Context, status string, userID string) ([]*domain.Task, error)
    Update(ctx context.Context, task *domain.Task) error
    Delete(ctx context.Context, id string) error
    BulkUpdateStatus(ctx context.Context, ids []string, status string) error
    Search(ctx context.Context, query string, userID string) ([]*domain.Task, error)
}

// UserRepository defines all user-related database operations
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id string) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    GetByUsername(ctx context.Context, username string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}

// TagRepository defines all tag-related database operations
type TagRepository interface {
    Create(ctx context.Context, tag *domain.Tag) error
    GetByID(ctx context.Context, id string) (*domain.Tag, error)
    GetByName(ctx context.Context, name string) (*domain.Tag, error)
    GetAll(ctx context.Context) ([]*domain.Tag, error)
    Update(ctx context.Context, tag *domain.Tag) error
    Delete(ctx context.Context, id string) error
    AddToTask(ctx context.Context, taskID string, tagID string) error
    RemoveFromTask(ctx context.Context, taskID string, tagID string) error
    GetTaskTags(ctx context.Context, taskID string) ([]*domain.Tag, error)
}