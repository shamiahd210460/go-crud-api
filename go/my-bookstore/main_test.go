package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupRouter sets up the Gin router with the routes for testing
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/books", getBooks)
	return router
}

func TestGetBooks(t *testing.T) {
	// Load sample data for testing
	books = []Book{
		{ID: "1", Title: "Test Book 1", Description: "A test book description 1", Price: 9.99, Quantity: 10},
		{ID: "2", Title: "Test Book 2", Description: "A test book description 2", Price: 12.99, Quantity: 5},
	}

	router := setupRouter()

	// Create a request to pass to our handler
	req, _ := http.NewRequest("GET", "/books", nil)
	// Create a response recorder to record the response
	w := httptest.NewRecorder()
	// Serve the HTTP request
	router.ServeHTTP(w, req)

	// Check the status code is what we expect
	assert.Equal(t, http.StatusOK, w.Code)

	// Check the response body is what we expect
	var response []Book
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "Test Book 1", response[0].Title)
	assert.Equal(t, "Test Book 2", response[1].Title)
}
