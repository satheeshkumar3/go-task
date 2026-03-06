package usecase

import (
    "context"
    "errors"
    "time"

    "task-manager/internal/domain"
    "task-manager/internal/repository/interfaces"
    "task-manager/internal/pkg/validator"
)

type TaskUseCase struct {
    taskRepo interfaces.TaskRepository
    userRepo interfaces.UserRepository
    tagRepo  interfaces.TagRepository
}

func NewTaskUseCase(
    taskRepo interfaces.TaskRepository,
    userRepo interfaces.UserRepository,
    tagRepo interfaces.TagRepository,
) *TaskUseCase {
    return &TaskUseCase{
        taskRepo: taskRepo,
        userRepo: userRepo,
        tagRepo:  tagRepo,
    }
}

// CreateTaskInput represents the input for creating a task
type CreateTaskInput struct {
    Title       string    `json:"title" validate:"required,min=3,max=100"`
    Description string    `json:"description" validate:"max=1000"`
    Priority    string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
    DueDate     *time.Time `json:"due_date"`
    UserID      string    `json:"user_id" validate:"required"`
    TagNames    []string  `json:"tag_names"`
}

// UpdateTaskInput represents the input for updating a task
type UpdateTaskInput struct {
    Title       string    `json:"title" validate:"omitempty,min=3,max=100"`
    Description string    `json:"description" validate:"omitempty,max=1000"`
    Status      string    `json:"status" validate:"omitempty,oneof=pending in_progress completed archived"`
    Priority    string    `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
    DueDate     *time.Time `json:"due_date"`
}

// TaskFilter represents filters for listing tasks
type TaskFilter struct {
    Status   string
    Priority string
    Page     int
    Limit    int
    UserID   string
}

// Create creates a new task
func (uc *TaskUseCase) Create(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
    // Validate input
    if err := validator.Validate(input); err != nil {
        return nil, err
    }

    // Check if user exists
    user, err := uc.userRepo.GetByID(ctx, input.UserID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not found")
    }

    // Set default priority if not provided
    if input.Priority == "" {
        input.Priority = domain.PriorityMedium
    }

    // Create task
    task := &domain.Task{
        Title:       input.Title,
        Description: input.Description,
        Status:      domain.StatusPending,
        Priority:    input.Priority,
        DueDate:     input.DueDate,
        UserID:      input.UserID,
    }

    if err := uc.taskRepo.Create(ctx, task); err != nil {
        return nil, err
    }

    // Add tags if provided
    for _, tagName := range input.TagNames {
        tag, err := uc.tagRepo.GetByName(ctx, tagName)
        if err != nil {
            continue
        }
        if tag != nil {
            uc.tagRepo.AddToTask(ctx, task.ID, tag.ID)
        }
    }

    return task, nil
}

// GetByID retrieves a task by ID
func (uc *TaskUseCase) GetByID(ctx context.Context, id string) (*domain.Task, error) {
    task, err := uc.taskRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if task == nil {
        return nil, errors.New("task not found")
    }
    return task, nil
}

// List retrieves tasks with filtering and pagination
func (uc *TaskUseCase) List(ctx context.Context, filter TaskFilter) ([]*domain.Task, int64, error) {
    // Set defaults
    if filter.Page < 1 {
        filter.Page = 1
    }
    if filter.Limit < 1 {
        filter.Limit = 10
    }
    if filter.Limit > 100 {
        filter.Limit = 100
    }

    if filter.Status != "" {
        return uc.taskRepo.GetByStatus(ctx, filter.Status, filter.UserID)
    }

    return uc.taskRepo.GetByUserID(ctx, filter.UserID, filter.Page, filter.Limit)
}

// Update updates an existing task
func (uc *TaskUseCase) Update(ctx context.Context, id string, input UpdateTaskInput) (*domain.Task, error) {
    // Validate input
    if err := validator.Validate(input); err != nil {
        return nil, err
    }

    // Get existing task
    task, err := uc.taskRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if task == nil {
        return nil, errors.New("task not found")
    }

    // Update fields
    if input.Title != "" {
        task.Title = input.Title
    }
    if input.Description != "" {
        task.Description = input.Description
    }
    if input.Status != "" {
        task.Status = input.Status
    }
    if input.Priority != "" {
        task.Priority = input.Priority
    }
    if input.DueDate != nil {
        task.DueDate = input.DueDate
    }

    // Save
    if err := uc.taskRepo.Update(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}

// Delete deletes a task
func (uc *TaskUseCase) Delete(ctx context.Context, id string) error {
    task, err := uc.taskRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    if task == nil {
        return errors.New("task not found")
    }

    return uc.taskRepo.Delete(ctx, id)
}

// BulkUpdateStatus updates status for multiple tasks
func (uc *TaskUseCase) BulkUpdateStatus(ctx context.Context, ids []string, status string) error {
    return uc.taskRepo.BulkUpdateStatus(ctx, ids, status)
}

// Search searches tasks by query
func (uc *TaskUseCase) Search(ctx context.Context, query string, userID string) ([]*domain.Task, error) {
    return uc.taskRepo.Search(ctx, query, userID)
}