package main

import (
    "log"
    "os"

    "task-manager/configs"
    "task-manager/internal/domain"
    "task-manager/internal/pkg/database"

    "github.com/joho/godotenv"
    "gorm.io/gorm"
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

    // Run migrations
    log.Println("Running migrations...")

    if err := db.AutoMigrate(); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // Create extensions if needed
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
        log.Printf("Warning: Failed to create uuid-ossp extension: %v", err)
    }

    // Create indexes
    log.Println("Creating indexes...")
    
    // Task indexes
    if err := db.Model(&domain.Task{}).AddIndex("idx_tasks_user_id", "user_id").Error; err != nil {
        log.Printf("Warning: Failed to create index: %v", err)
    }
    
    if err := db.Model(&domain.Task{}).AddIndex("idx_tasks_status", "status").Error; err != nil {
        log.Printf("Warning: Failed to create index: %v", err)
    }
    
    if err := db.Model(&domain.Task{}).AddIndex("idx_tasks_due_date", "due_date").Error; err != nil {
        log.Printf("Warning: Failed to create index: %v", err)
    }

    // User indexes
    if err := db.Model(&domain.User{}).AddIndex("idx_users_email", "email").Error; err != nil {
        log.Printf("Warning: Failed to create index: %v", err)
    }
    
    if err := db.Model(&domain.User{}).AddIndex("idx_users_username", "username").Error; err != nil {
        log.Printf("Warning: Failed to create index: %v", err)
    }

    // Tag indexes
    if err := db.Model(&domain.Tag{}).AddIndex("idx_tags_name", "name").Error; err != nil {
        log.Printf("Warning: Failed to create index: %v", err)
    }

    // Create full-text search indexes
    if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_tasks_search 
        ON tasks USING GIN (to_tsvector('english', title || ' ' || COALESCE(description, '')))
    `).Error; err != nil {
        log.Printf("Warning: Failed to create full-text search index: %v", err)
    }

    log.Println("Migrations completed successfully")
}