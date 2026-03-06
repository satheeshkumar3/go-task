package handler

import (
    "net/http"
    "strconv"

    "task-manager/internal/usecase"

    "github.com/gin-gonic/gin"
)

type TaskHandler struct {
    taskUseCase *usecase.TaskUseCase
}

func NewTaskHandler(taskUseCase *usecase.TaskUseCase) *TaskHandler {
    return &TaskHandler{
        taskUseCase: taskUseCase,
    }
}

// CreateTask handles POST /tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
    var input usecase.CreateTaskInput
    
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get user ID from context (set by auth middleware)
    userID, exists := c.Get("user_id")
    if exists {
        input.UserID = userID.(string)
    }

    task, err := h.taskUseCase.Create(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, task)
}

// GetTask handles GET /tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
    id := c.Param("id")
    
    task, err := h.taskUseCase.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, task)
}

// ListTasks handles GET /tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
    // Get user ID from context
    userID, _ := c.Get("user_id")
    
    // Parse query parameters
    status := c.Query("status")
    priority := c.Query("priority")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    filter := usecase.TaskFilter{
        Status:   status,
        Priority: priority,
        Page:     page,
        Limit:    limit,
        UserID:   userID.(string),
    }

    tasks, total, err := h.taskUseCase.List(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":  tasks,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}

// UpdateTask handles PUT /tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    id := c.Param("id")
    
    var input usecase.UpdateTaskInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    task, err := h.taskUseCase.Update(c.Request.Context(), id, input)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, task)
}

// DeleteTask handles DELETE /tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
    id := c.Param("id")
    
    if err := h.taskUseCase.Delete(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}

// BulkUpdateStatus handles PATCH /tasks/bulk/status
func (h *TaskHandler) BulkUpdateStatus(c *gin.Context) {
    var input struct {
        IDs    []string `json:"ids" binding:"required"`
        Status string   `json:"status" binding:"required,oneof=pending in_progress completed archived"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.taskUseCase.BulkUpdateStatus(c.Request.Context(), input.IDs, input.Status); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Tasks updated successfully"})
}

// SearchTasks handles GET /tasks/search
func (h *TaskHandler) SearchTasks(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
        return
    }

    userID, _ := c.Get("user_id")

    tasks, err := h.taskUseCase.Search(c.Request.Context(), query, userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, tasks)
}