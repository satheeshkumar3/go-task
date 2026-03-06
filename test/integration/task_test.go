//go:build integration
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "task-manager/configs"
    "task-manager/internal/delivery/http/handler"
    "task-manager/internal/delivery/http/router"
    "task-manager/internal/pkg/database"
    "task-manager/internal/repository/postgres"
    "task-manager/internal/usecase"
    "task-manager/pkg/middleware"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "golang.org/x/time/rate"
)

func setupTestDB(t *testing.T) *database.Database {
    config := &configs.DatabaseConfig{
        Host:         "localhost",
        Port:         "5432",
        User:         "postgres",
        Password:     "postgres",
        DBName:       "taskmanager_test",
        SSLMode:      "disable",
        MaxOpenConns: 10,
        MaxIdleConns: 5,
        MaxLifetime:  5,
    }

    db, err := database.NewPostgresConnection(config)
    require.NoError(t, err)

    // Clean up database
    err = db.Exec("TRUNCATE TABLE tasks, users, tags RESTART IDENTITY CASCADE").Error
    require.NoError(t, err)

    return db
}

func setupTestServer(t *testing.T) (*gin.Engine, *database.Database) {
    // Set Gin to test mode
    gin.SetMode(gin.TestMode)

    // Setup database
    db := setupTestDB(t)

    // Initialize repositories
    taskRepo := postgres.NewTaskRepository(db.DB)
    userRepo := postgres.NewUserRepository(db.DB)
    tagRepo := postgres.NewTagRepository(db.DB)

    // Initialize use cases
    taskUseCase := usecase.NewTaskUseCase(taskRepo, userRepo, tagRepo)
    authUseCase := usecase.NewAuthUseCase(userRepo, "test-secret", 24)

    // Initialize handlers
    taskHandler := handler.NewTaskHandler(taskUseCase)
    authHandler := handler.NewAuthHandler(authUseCase)

    // Initialize middleware
    authMW := middleware.NewAuthMiddleware("test-secret")

    // Initialize router
    routerConfig := &router.Config{
        JWTSecret:       "test-secret",
        RateLimit:       rate.Limit(100),
        RateLimitBurst:  100,
        Environment:     "test",
    }

    r := router.NewRouter(
        taskHandler,
        authHandler,
        authMW,
        routerConfig,
    )

    return r.Engine(), db
}

func createTestUser(t *testing.T, router *gin.Engine) (string, string) {
    // Register user
    registerBody := map[string]string{
        "email":    "test@example.com",
        "username": "testuser",
        "password": "password123",
    }
    
    body, _ := json.Marshal(registerBody)
    req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    require.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    
    accessToken := response["access_token"].(string)
    user := response["user"].(map[string]interface{})
    userID := user["id"].(string)
    
    return accessToken, userID
}

func TestTaskCRUD(t *testing.T) {
    router, db := setupTestServer(t)
    defer db.DB.DB()

    // Create test user and get token
    token, userID := createTestUser(t, router)

    t.Run("Create Task", func(t *testing.T) {
        taskData := map[string]interface{}{
            "title":       "Integration Test Task",
            "description": "This is a test task",
            "priority":    "high",
        }
        
        body, _ := json.Marshal(taskData)
        req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", "Bearer "+token)
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var createdTask map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &createdTask)
        assert.NoError(t, err)
        
        assert.Equal(t, taskData["title"], createdTask["title"])
        assert.Equal(t, taskData["description"], createdTask["description"])
        assert.Equal(t, taskData["priority"], createdTask["priority"])
        assert.Equal(t, "pending", createdTask["status"])
        assert.Equal(t, userID, createdTask["user_id"])
        assert.NotEmpty(t, createdTask["id"])
    })

    t.Run("List Tasks", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/tasks?page=1&limit=10", nil)
        req.Header.Set("Authorization", "Bearer "+token)
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        
        data := response["data"].([]interface{})
        assert.GreaterOrEqual(t, len(data), 1)
        assert.GreaterOrEqual(t, response["total"].(float64), 1.0)
    })

    t.Run("Unauthorized Access", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}

func TestAuthentication(t *testing.T) {
    router, db := setupTestServer(t)
    defer db.DB.DB()

    t.Run("Register", func(t *testing.T) {
        body := map[string]string{
            "email":    "newuser@example.com",
            "username": "newuser",
            "password": "password123",
        }
        
        jsonBody, _ := json.Marshal(body)
        req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        
        assert.NotEmpty(t, response["access_token"])
        assert.NotEmpty(t, response["refresh_token"])
        
        user := response["user"].(map[string]interface{})
        assert.Equal(t, body["email"], user["email"])
        assert.Equal(t, body["username"], user["username"])
    })

    t.Run("Login", func(t *testing.T) {
        // First register
        body := map[string]string{
            "email":    "login@example.com",
            "username": "loginuser",
            "password": "password123",
        }
        
        jsonBody, _ := json.Marshal(body)
        req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        assert.Equal(t, http.StatusCreated, w.Code)
        
        // Then login
        loginBody := map[string]string{
            "email":    "login@example.com",
            "password": "password123",
        }
        
        jsonBody, _ = json.Marshal(loginBody)
        req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")
        
        w = httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        
        assert.NotEmpty(t, response["access_token"])
    })

    t.Run("Invalid Login", func(t *testing.T) {
        loginBody := map[string]string{
            "email":    "nonexistent@example.com",
            "password": "wrongpassword",
        }
        
        jsonBody, _ := json.Marshal(loginBody)
        req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}