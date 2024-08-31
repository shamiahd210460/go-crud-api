package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Book struct defines the structure for the book information
type Book struct {
	ID              string  `json:"ID"`
	BookId          string  `json:"bookId"`
	AuthorId        string  `json:"authorId"`
	PublisherId     string  `json:"publisherId"`
	Title           string  `json:"title"`
	PublicationDate string  `json:"publicationDate"`
	Isbn            string  `json:"isbn"`
	Pages           int     `json:"pages"`
	Genre           string  `json:"genre"`
	Description     string  `json:"description"`
	Price           float64 `json:"price"`
	Quantity        int     `json:"quantity"`
}

var books []Book

func main() {
	loadBooksFromJSON("books.json")

	router := gin.Default()
	router.Use(corsMiddleware())
	registerRoutes(router)
	router.Run("0.0.0.0:8081")
}

// Load books data from JSON file
func loadBooksFromJSON(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(bytes, &books); err != nil {
		panic(err)
	}
}

// Register all the routes
func registerRoutes(router *gin.Engine) {
	router.GET("/books", getBooks)
	router.POST("/books", postBooks)
	router.GET("/books/:ID", getBookByID)
	router.PUT("/books/:ID", updateBookByID)
	router.DELETE("/books/:ID", deleteBookByID)
	router.GET("/books/search", searchBooks)
}

// Retrieve all the books with pagination
func getBooks(c *gin.Context) {
	limit, offset, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	start := offset
	end := offset + limit

	if start > len(books) {
		start = len(books)
	}
	if end > len(books) {
		end = len(books)
	}

	paginatedBooks := books[start:end]
	c.IndentedJSON(http.StatusOK, paginatedBooks)
}

// Parse pagination parameters from the context
func parsePagination(c *gin.Context) (int, int, error) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 0, 0, fmt.Errorf("invalid limit")
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, fmt.Errorf("invalid offset")
	}

	return limit, offset, nil
}

// Add new book details
func postBooks(c *gin.Context) {
	var newBook Book

	if err := c.BindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book data"})
		return
	}

	books = append(books, newBook)
	c.IndentedJSON(http.StatusCreated, newBook)
}

// Get book by ID
func getBookByID(c *gin.Context) {
	ID := c.Param("ID")

	for _, book := range books {
		if book.ID == ID {
			c.IndentedJSON(http.StatusOK, book)
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

// Update book by ID
func updateBookByID(c *gin.Context) {
	ID := c.Param("ID")
	var updatedBook Book

	if err := c.BindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book data"})
		return
	}

	for i, book := range books {
		if book.ID == ID {
			books[i] = updatedBook
			c.IndentedJSON(http.StatusOK, updatedBook)
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

// Delete book by ID
func deleteBookByID(c *gin.Context) {
	ID := c.Param("ID")

	for i, book := range books {
		if book.ID == ID {
			books = append(books[:i], books[i+1:]...)
			c.IndentedJSON(http.StatusOK, gin.H{"message": "book deleted"})
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "book not found"})
}

// Search books by query
func searchBooks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "query parameter 'q' is required"})
		return
	}

	results := performSearch(query)
	c.IndentedJSON(http.StatusOK, results)
}

// Perform search with multiple workers
func performSearch(query string) []Book {
	query = strings.ToLower(query)
	results := make([]Book, 0)

	numWorkers := 4
	bookChannel := make(chan Book, len(books))
	resultChannel := make(chan Book)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go searchWorker(query, bookChannel, resultChannel, &wg)
	}

	// Send books to the bookChannel
	go func() {
		for _, book := range books {
			bookChannel <- book
		}
		close(bookChannel)
	}()

	// Close the result channel once all workers are done
	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	// Collect results from the result channel
	for result := range resultChannel {
		results = append(results, result)
	}

	return results
}

// Worker function to search books based on the query
func searchWorker(query string, books <-chan Book, results chan<- Book, wg *sync.WaitGroup) {
	defer wg.Done()
	for book := range books {
		if strings.Contains(strings.ToLower(book.Title), query) || strings.Contains(strings.ToLower(book.Description), query) {
			results <- book
		}
	}
}

// CORS middleware to handle cross-origin requests
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
