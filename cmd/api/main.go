package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "task-manager/configs"
    "task-manager/internal/delivery/http/handler"
    "task-manager/internal/delivery/http/router"
    "task-manager/internal/pkg/database"
    "task-manager/internal/repository/postgres"
    "task-manager/internal/usecase"
    "task-manager/pkg/middleware"

    "golang.org/x/time/rate"
)

func main() {
    // Load configuration
    config, err := configs.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database
    db, err := database.NewPostgresConnection(&config.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Run migrations
    if err := db.AutoMigrate(); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // Initialize repositories
    taskRepo := postgres.NewTaskRepository(db.DB)
    userRepo := postgres.NewUserRepository(db.DB)
    tagRepo := postgres.NewTagRepository(db.DB)

    // Initialize use cases
    taskUseCase := usecase.NewTaskUseCase(taskRepo, userRepo, tagRepo)
    authUseCase := usecase.NewAuthUseCase(userRepo, config.Auth.JWTSecret, config.Auth.TokenDuration)

    // Initialize handlers
    taskHandler := handler.NewTaskHandler(taskUseCase)
    authHandler := handler.NewAuthHandler(authUseCase)

    // Initialize middleware
    authMiddleware := middleware.NewAuthMiddleware(config.Auth.JWTSecret)

    // Initialize router
    routerConfig := &router.Config{
        JWTSecret:      config.Auth.JWTSecret,
        RateLimit:      rate.Limit(10), // 10 requests per second
        RateLimitBurst: 20,              // Burst of 20
        Environment:    config.Server.Environment,
    }

    r := router.NewRouter(
        taskHandler,
        authHandler,
        authMiddleware,
        routerConfig,
    )

    // Create server
    srv := &http.Server{
        Addr:         ":" + config.Server.Port,
        Handler:      r.Engine(),
        ReadTimeout:  config.Server.ReadTimeout,
        WriteTimeout: config.Server.WriteTimeout,
        IdleTimeout:  config.Server.IdleTimeout,
    }

    // Start server in goroutine
    go func() {
        log.Printf("Server starting on port %s", config.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    // Close database connection
    sqlDB, err := db.DB.DB()
    if err == nil {
        sqlDB.Close()
    }

    log.Println("Server exited gracefully")
}