package middleware

import (
    "bytes"
    "io"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type bodyLogWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate request ID
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        // Read body
        var bodyBytes []byte
        if c.Request.Body != nil {
            bodyBytes, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        }

        // Start timer
        start := time.Now()

        // Create custom response writer
        blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
        c.Writer = blw

        // Process request
        c.Next()

        // Log after request
        latency := time.Since(start)

        // Log request details
        gin.DefaultWriter.Write([]byte(
            "\n" +
            "Request ID: " + requestID + "\n" +
            "Method: " + c.Request.Method + "\n" +
            "Path: " + c.Request.URL.Path + "\n" +
            "Status: " + c.Writer.Status() + "\n" +
            "Latency: " + latency.String() + "\n" +
            "Client IP: " + c.ClientIP() + "\n" +
            "Request Body: " + string(bodyBytes) + "\n" +
            "Response Body: " + blw.body.String() + "\n" +
            "----------------------------------------\n",
        ))
    }
}