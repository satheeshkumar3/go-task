package router

import (
    "task-manager/internal/delivery/http/handler"
    "task-manager/pkg/middleware"

    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

type Router struct {
    engine      *gin.Engine
    taskHandler *handler.TaskHandler
    authHandler *handler.AuthHandler
    authMW      *middleware.AuthMiddleware
    config      *Config
}

type Config struct {
    JWTSecret     string
    RateLimit     rate.Limit
    RateLimitBurst int
    Environment   string
}

func NewRouter(
    taskHandler *handler.TaskHandler,
    authHandler *handler.AuthHandler,
    authMW *middleware.AuthMiddleware,
    config *Config,
) *Router {
    // Set Gin mode
    if config.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := &Router{
        engine:      gin.New(),
        taskHandler: taskHandler,
        authHandler: authHandler,
        authMW:      authMW,
        config:      config,
    }

    router.setupMiddleware()
    router.setupRoutes()

    return router
}

func (r *Router) setupMiddleware() {
    // Global middleware
    r.engine.Use(middleware.Logger())
    r.engine.Use(middleware.Recovery())
    r.engine.Use(middleware.CORS())
    
    // Rate limiting
    rateLimiter := middleware.NewRateLimiter(r.config.RateLimit, r.config.RateLimitBurst)
    r.engine.Use(middleware.RateLimit(rateLimiter))
}

func (r *Router) setupRoutes() {
    // Health check
    r.engine.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API v1 group
    v1 := r.engine.Group("/api/v1")
    {
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", r.authHandler.Register)
            auth.POST("/login", r.authHandler.Login)
            auth.POST("/refresh", r.authHandler.RefreshToken)
        }

        // Protected routes
        protected := v1.Group("/")
        protected.Use(r.authMW.RequireAuth())
        {
            // Tasks
            tasks := protected.Group("/tasks")
            {
                tasks.POST("/", r.taskHandler.CreateTask)
                tasks.GET("/", r.taskHandler.ListTasks)
                tasks.GET("/search", r.taskHandler.SearchTasks)
                tasks.PATCH("/bulk/status", r.taskHandler.BulkUpdateStatus)
                tasks.GET("/:id", r.taskHandler.GetTask)
                tasks.PUT("/:id", r.taskHandler.UpdateTask)
                tasks.DELETE("/:id", r.taskHandler.DeleteTask)
            }

            // Users
            users := protected.Group("/users")
            {
                users.GET("/profile", r.authHandler.GetProfile)
                users.PUT("/profile", r.authHandler.UpdateProfile)
            }

            // Tags
            tags := protected.Group("/tags")
            {
                tags.GET("/", r.authHandler.GetTags) // You'll need to implement this
                tags.POST("/", r.authHandler.CreateTag)
                tags.DELETE("/:id", r.authHandler.DeleteTag)
            }
        }
    }
}

func (r *Router) Engine() *gin.Engine {
    return r.engine
}