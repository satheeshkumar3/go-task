package handler

import (
    "net/http"
    "time"

    "task-manager/internal/usecase"

    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
    return &AuthHandler{
        authUseCase: authUseCase,
    }
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
    User         gin.H     `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.authUseCase.Register(c.Request.Context(), usecase.RegisterInput{
        Email:    req.Email,
        Username: req.Username,
        Password: req.Password,
    })
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, AuthResponse{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
        ExpiresAt:    result.ExpiresAt,
        User: gin.H{
            "id":       result.User.ID,
            "email":    result.User.Email,
            "username": result.User.Username,
        },
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.authUseCase.Login(c.Request.Context(), usecase.LoginInput{
        Email:    req.Email,
        Password: req.Password,
    })
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    c.JSON(http.StatusOK, AuthResponse{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
        ExpiresAt:    result.ExpiresAt,
        User: gin.H{
            "id":       result.User.ID,
            "email":    result.User.Email,
            "username": result.User.Username,
        },
    })
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
        return
    }

    c.JSON(http.StatusOK, AuthResponse{
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
        ExpiresAt:    result.ExpiresAt,
        User: gin.H{
            "id":       result.User.ID,
            "email":    result.User.Email,
            "username": result.User.Username,
        },
    })
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    user, err := h.authUseCase.GetUserByID(c.Request.Context(), userID.(string))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":        user.ID,
        "email":     user.Email,
        "username":  user.Username,
        "created_at": user.CreatedAt,
    })
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    var input usecase.UpdateProfileInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.authUseCase.UpdateProfile(c.Request.Context(), userID.(string), input)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":        user.ID,
        "email":     user.Email,
        "username":  user.Username,
        "updated_at": user.UpdatedAt,
    })
}