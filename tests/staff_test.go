package controllers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "agnos-hospital-middleware/controllers"
    "agnos-hospital-middleware/config"
    "agnos-hospital-middleware/models"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestCreateStaff(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    config.InitDB()

    r := gin.Default()
    r.POST("/staff/create", controllers.CreateStaff)

    body := map[string]string{
        "username": "testuser",
        "password": "password123",
        "hospital": "Hospital A",
    }
    jsonBody, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", "/staff/create", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "Staff created successfully")
}
