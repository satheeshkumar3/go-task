package main

import (
    "log"
    "time"

    "task-manager/configs"
    "task-manager/internal/domain"
    "task-manager/internal/pkg/database"

    "github.com/joho/godotenv"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Load configuration
    config, err := configs.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Connect to database
    db, err := database.NewPostgresConnection(&config.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    log.Println("Seeding database...")

    // Create users
    users := createUsers()
    for _, user := range users {
        if err := db.Create(user).Error; err != nil {
            log.Printf("Failed to create user %s: %v", user.Email, err)
        }
    }

    // Create tags
    tags := createTags()
    for _, tag := range tags {
        if err := db.Create(tag).Error; err != nil {
            log.Printf("Failed to create tag %s: %v", tag.Name, err)
        }
    }

    // Create tasks for each user
    for _, user := range users {
        tasks := createTasks(user.ID)
        for _, task := range tasks {
            if err := db.Create(task).Error; err != nil {
                log.Printf("Failed to create task for user %s: %v", user.ID, err)
            }
        }
    }

    // Associate tags with tasks
    var allTasks []domain.Task
    db.Find(&allTasks)
    
    var allTags []domain.Tag
    db.Find(&allTags)

    if len(allTags) > 0 {
        for i, task := range allTasks {
            // Add 1-3 random tags to each task
            numTags := i%3 + 1
            for j := 0; j < numTags && j < len(allTags); j++ {
                tagIndex := (i + j) % len(allTags)
                db.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
                    task.ID, allTags[tagIndex].ID)
            }
        }
    }

    log.Println("Seeding completed successfully")
}

func createUsers() []*domain.User {
    password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    
    return []*domain.User{
        {
            Email:    "john@example.com",
            Username: "john_doe",
            Password: string(password),
        },
        {
            Email:    "jane@example.com",
            Username: "jane_smith",
            Password: string(password),
        },
        {
            Email:    "bob@example.com",
            Username: "bob_wilson",
            Password: string(password),
        },
    }
}

func createTags() []*domain.Tag {
    return []*domain.Tag{
        {Name: "urgent", Color: "#FF0000"},
        {Name: "work", Color: "#0000FF"},
        {Name: "personal", Color: "#00FF00"},
        {Name: "meeting", Color: "#FFA500"},
        {Name: "deadline", Color: "#800080"},
        {Name: "idea", Color: "#FFC0CB"},
    }
}

func createTasks(userID string) []*domain.Task {
    now := time.Now()
    
    return []*domain.Task{
        {
            Title:       "Complete project proposal",
            Description: "Write and submit the Q1 project proposal",
            Status:      domain.StatusInProgress,
            Priority:    domain.PriorityHigh,
            DueDate:     ptrTime(now.AddDate(0, 0, 3)),
            UserID:      userID,
        },
        {
            Title:       "Review pull requests",
            Description: "Review team members' PRs and provide feedback",
            Status:      domain.StatusPending,
            Priority:    domain.PriorityMedium,
            DueDate:     ptrTime(now.AddDate(0, 0, 1)),
            UserID:      userID,
        },
        {
            Title:       "Update documentation",
            Description: "Update API documentation with recent changes",
            Status:      domain.StatusCompleted,
            Priority:    domain.PriorityLow,
            DueDate:     ptrTime(now.AddDate(0, 0, -2)),
            UserID:      userID,
        },
        {
            Title:       "Team meeting",
            Description: "Weekly sync with the development team",
            Status:      domain.StatusCompleted,
            Priority:    domain.PriorityMedium,
            DueDate:     ptrTime(now.AddDate(0, 0, -1)),
            UserID:      userID,
        },
        {
            Title:       "Deploy to staging",
            Description: "Deploy latest version to staging environment",
            Status:      domain.StatusPending,
            Priority:    domain.PriorityUrgent,
            DueDate:     ptrTime(now.AddDate(0, 0, 2)),
            UserID:      userID,
        },
    }
}

func ptrTime(t time.Time) *time.Time {
    return &t
}