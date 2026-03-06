package domain

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// Task represents the core business entity
type Task struct {
    ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Title       string     `gorm:"type:varchar(100);not null;index" json:"title"`
    Description string     `gorm:"type:text" json:"description"`
    Status      string     `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
    Priority    string     `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
    DueDate     *time.Time `json:"due_date,omitempty"`
    UserID      string     `gorm:"type:uuid;not null;index" json:"user_id"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
    
    // Relations
    User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Tags        []Tag      `gorm:"many2many:task_tags;" json:"tags,omitempty"`
}

// User represents a user entity
type User struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
    Username  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
    Password  string    `gorm:"type:varchar(100);not null" json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Tag represents a tag for tasks
type Tag struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Name      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
    Color     string    `gorm:"type:varchar(7);default:'#000000'" json:"color"`
    CreatedAt time.Time `json:"created_at"`
}

// TaskStatus constants
const (
    StatusPending    = "pending"
    StatusInProgress = "in_progress"
    StatusCompleted  = "completed"
    StatusArchived   = "archived"
)

// TaskPriority constants
const (
    PriorityLow    = "low"
    PriorityMedium = "medium"
    PriorityHigh   = "high"
    PriorityUrgent = "urgent"
)

// BeforeCreate GORM hook
func (t *Task) BeforeCreate(tx *gorm.DB) error {
    if t.ID == "" {
        t.ID = uuid.New().String()
    }
    return nil
}

// BeforeCreate GORM hook for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    return nil
}

// BeforeCreate GORM hook for Tag
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
    if t.ID == "" {
        t.ID = uuid.New().String()
    }
    return nil
}