package postgres

import (
    "context"
    "errors"
    "time"

    "task-manager/internal/domain"
    "task-manager/internal/repository/interfaces"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type taskRepository struct {
    db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) interfaces.TaskRepository {
    return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
    return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
    var task domain.Task
    err := r.db.WithContext(ctx).
        Preload("User").
        Preload("Tags").
        Where("id = ?", id).
        First(&task).Error
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &task, err
}

func (r *taskRepository) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Task, int64, error) {
    var tasks []*domain.Task
    var total int64

    offset := (page - 1) * limit

    // Get total count
    if err := r.db.WithContext(ctx).Model(&domain.Task{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Get paginated results
    err := r.db.WithContext(ctx).
        Preload("Tags").
        Where("user_id = ?", userID).
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&tasks).Error

    return tasks, total, err
}

func (r *taskRepository) GetByStatus(ctx context.Context, status string, userID string) ([]*domain.Task, error) {
    var tasks []*domain.Task
    err := r.db.WithContext(ctx).
        Preload("Tags").
        Where("status = ? AND user_id = ?", status, userID).
        Order("due_date ASC NULLS LAST").
        Find(&tasks).Error
    return tasks, err
}

func (r *taskRepository) Update(ctx context.Context, task *domain.Task) error {
    task.UpdatedAt = time.Now()
    return r.db.WithContext(ctx).Save(task).Error
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&domain.Task{}, "id = ?", id).Error
}

func (r *taskRepository) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
    return r.db.WithContext(ctx).Model(&domain.Task{}).
        Where("id IN ?", ids).
        Updates(map[string]interface{}{
            "status": status,
            "updated_at": time.Now(),
        }).Error
}

func (r *taskRepository) Search(ctx context.Context, query string, userID string) ([]*domain.Task, error) {
    var tasks []*domain.Task
    
    // Full-text search using PostgreSQL
    err := r.db.WithContext(ctx).
        Preload("Tags").
        Where("user_id = ?", userID).
        Where("to_tsvector('english', title || ' ' || COALESCE(description, '')) @@ to_tsquery('english', ?)", query).
        Order("ts_rank(to_tsvector('english', title || ' ' || COALESCE(description, '')), to_tsquery('english', ?)) DESC", query).
        Find(&tasks).Error
    
    return tasks, err
}

// UserRepository implementation
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
    user.UpdatedAt = time.Now()
    return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// TagRepository implementation
type tagRepository struct {
    db *gorm.DB
}

func NewTagRepository(db *gorm.DB) interfaces.TagRepository {
    return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) error {
    return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
    var tag domain.Tag
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &tag, err
}

func (r *tagRepository) GetByName(ctx context.Context, name string) (*domain.Tag, error) {
    var tag domain.Tag
    err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &tag, err
}

func (r *tagRepository) GetAll(ctx context.Context) ([]*domain.Tag, error) {
    var tags []*domain.Tag
    err := r.db.WithContext(ctx).Order("name").Find(&tags).Error
    return tags, err
}

func (r *tagRepository) Update(ctx context.Context, tag *domain.Tag) error {
    return r.db.WithContext(ctx).Save(tag).Error
}

func (r *tagRepository) Delete(ctx context.Context, id string) error {
    // Remove associations first
    if err := r.db.WithContext(ctx).Exec("DELETE FROM task_tags WHERE tag_id = ?", id).Error; err != nil {
        return err
    }
    return r.db.WithContext(ctx).Delete(&domain.Tag{}, "id = ?", id).Error
}

func (r *tagRepository) AddToTask(ctx context.Context, taskID string, tagID string) error {
    return r.db.WithContext(ctx).Exec(
        "INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
        taskID, tagID,
    ).Error
}

func (r *tagRepository) RemoveFromTask(ctx context.Context, taskID string, tagID string) error {
    return r.db.WithContext(ctx).Exec(
        "DELETE FROM task_tags WHERE task_id = ? AND tag_id = ?",
        taskID, tagID,
    ).Error
}

func (r *tagRepository) GetTaskTags(ctx context.Context, taskID string) ([]*domain.Tag, error) {
    var tags []*domain.Tag
    err := r.db.WithContext(ctx).
        Joins("JOIN task_tags ON task_tags.tag_id = tags.id").
        Where("task_tags.task_id = ?", taskID).
        Find(&tags).Error
    return tags, err
}