package middleware

import (
    "log"
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Log the error and stack trace
                log.Printf("PANIC: %v\n%s", err, debug.Stack())
                
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Internal server error",
                    "request_id": c.GetString("request_id"),
                })
                c.Abort()
            }
        }()
        c.Next()
    }
}