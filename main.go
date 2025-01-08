package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/rand"
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

	log.Println("Starting server on port 8083...")
	r.Run(":8083")
}

func response(c *gin.Context) {
	ctx := c.Request.Context()
	var q string = `
		select 
			json_agg(
				json_build_object(
					'data', data,
					'status', status
				)
			) as results
		from apis
		where url = $1 and method = $2
	`
	var jsonData []byte
	s := c.Query("muckStatus")
	var err error
	fmt.Println("c.Request.URL.Path")
	fmt.Println(c.Request.URL.Path)

	if s != "" {
		q += " and status = $3"
		err = db.QueryRow(ctx, q, c.Request.URL.Path, c.Request.Method, s).Scan(&jsonData)
	} else {
		err = db.QueryRow(ctx, q, c.Request.URL.Path, c.Request.Method).Scan(&jsonData)
	}

	if err != nil || len(jsonData) == 0 {
		log.Printf("not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var data []Result
	err = json.Unmarshal(jsonData, &data)

	if err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal JSON"})
		return
	}

	randomResult := rand.Intn(len(data))

	if data[randomResult].Data == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"message": "The requested URL was not found on this server",
		})
		return
	}

	var j map[string]interface{}
	if err = json.Unmarshal([]byte(data[randomResult].Data), &j); err == nil {
		c.JSON(data[randomResult].Status, j)
		return
	} else {
		c.JSON(data[randomResult].Status, data[randomResult].Data)
		return
	}

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
		"INSERT INTO apis (url, data, status, method) VALUES ($1, $2, $3, $4)",
		api.URL, string(dataJSON), api.Status, api.Method)

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
