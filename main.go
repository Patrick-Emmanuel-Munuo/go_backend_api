package main

import (
	"database/sql"

	"fmt"
	"log"
	"net/http"
	"os"
	"vartrick-server/controllers"
	"vartrick-server/route"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// This assumes a global DB variable
var db *sql.DB

type Options struct {
	Table       string                 `json:"table"`
	Select      []string               `json:"select"`
	Condition   map[string]interface{} `json:"condition"`
	OrCondition map[string]interface{} `json:"or_condition"`
}
type ReadResult struct {
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: No .env file found or failed to load")
	}
	// Recover from panic (simulating try-catch)
	defer func() {
		if r := recover(); r != nil {
			log.Printf(`{"success": false, "message": "Unexpected fatal error", "error": "%v"}`, r)
			os.Exit(1)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "2010" // Default port
	}

	router := gin.Default()

	// Initialize and check the DB connection with JSON logging
	// Initialize DB connection
	if err := InitDBConnection(); err != nil {
		log.Printf(`{"success": false, "message": "Failed to connect to MySQL", "error": "%v"}`, err)
		os.Exit(1)
	} else {
		log.Printf(`{"success": true, "message": "Connected to MySQL successfully"}`)
	}
	// Inject DB into controller package
	controllers.SetDB(db)
	// Basic welcome route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Golang Welcome to VarTrick Server application",
		})
	})
	// Use POST for Read because you expect JSON body
	// Load app routes
	route.SetupRouter(router)
	// Handle 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Wrong route or HTTP method",
		})
	})
	// Start the server
	if err := router.Run(":" + port); err != nil {
		log.Printf(`{"success": false, "message": "Failed to start Gin server", "error": "%v"}`, err)
		os.Exit(1)
	}
}

// InitDBConnection establishes and checks the MySQL connection
func InitDBConnection() error {
	databaseHost := os.Getenv("DATABASE_HOST")
	databaseUser := os.Getenv("DATABASE_USER")
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_NAME")
	//fmt.Println("DB user:", databaseUser)
	//fmt.Println("DB password:", databasePassword)
	//fmt.Println("DB host:", databaseHost)
	//fmt.Println("DB name:", databaseName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		databaseUser, databasePassword, databaseHost, databaseName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	return nil
}
