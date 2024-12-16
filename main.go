package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *pgxpool.Pool

func main() {

	log.Println("Starting database connection...")
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://postgres:1234@localhost:5432/postgres")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established.")

	r := gin.Default()

	r.POST("/muck/create", create)
	r.DELETE("/muck/:url", delete)
	r.NoRoute(response)

	log.Println("Starting server on port 8080...")
	r.Run(":8080")
}

func response(c *gin.Context) {
	fmt.Println("path, and method:", c.Request.URL.Path, c.Request.Method)
	ctx := c.Request.Context()
	var data string
	err := db.QueryRow(ctx, "select data from apis where url = $1", c.Request.URL.Path).Scan(&data)
	if err != nil || data == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"message": "The requested URL was not found on this server",
		})
		return
	}

	// Try to parse `data` as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err == nil {
		// If parsing succeeds, return the parsed JSON
		c.JSON(http.StatusOK, jsonData)
		return
	}

	// If parsing fails, return `data` as a plain string
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "Internal server error",
		"message": "An internal server error occurred",
	})
}

func create(c *gin.Context) {
	log.Println("Creating a new user...")
	ctx := c.Request.Context()
	var api API
	if err := c.ShouldBindJSON(&api); err != nil {
		log.Printf("Invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	dataJSON, err := json.Marshal(api.Data)
	if err != nil {
		log.Printf("Error marshalling data to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process data"})
		return
	}

	_, err = db.Exec(ctx,
		"INSERT INTO apis (url, data) VALUES ($1, $2)",
		api.URL, string(dataJSON))

	if err != nil {
		log.Printf("Error creating api: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create api"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "API created successfully"})
}

func delete(c *gin.Context) {
	url := c.Param("url")
	ctx := c.Request.Context()
	log.Printf("Deleting user with ID: %s", url)

	result, err := db.Exec(ctx, "DELETE FROM apis WHERE url=$1", "/"+url)

	if err != nil || result.RowsAffected() == 0 {
		log.Printf("Error deleting user with ID %s: %v", url, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	log.Printf("Api with ID %s deleted successfully.", url)
	c.JSON(http.StatusOK, gin.H{"message": "Api deleted successfully"})
}
