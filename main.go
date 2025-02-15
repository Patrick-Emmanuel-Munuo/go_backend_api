package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"vartrick-server/route"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "2080" // default port
	}
	router := gin.Default()
	//gin.SetMode(gin.ReleaseMode)
	// Middleware
	//router.Use(cors.Default())
	//router.Use(gin.Logger())
	//router.Use(helmet.Default()) "github.com/gin-contrib/helmet"

	// Rate limiting
	//rate := limiter.Rate{
	///	Period: 5 * time.Second,
	//	Limit:  200,
	//}
	//store := memory.NewStore()
	//limiterMiddleware := mgin.NewMiddleware(limiter.New(store, rate))
	//router.Use(limiterMiddleware)
	// // SetupMySQL initializes MySQL connection
	// Initialize and check the DB connection
	CheckDBConnection()
	// Routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Golang Welcome to VarTrick Server application",
		})
	})
	//Call mysqlrouter to setup MySQL-related routes
	route.SetupRouter(router)

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"message": "wrong router and api method ...",
		})
	})

	// Start the server
	router.Run(":" + port)
}

var db *sql.DB

// InitDBConnection function to establish and check MySQL connection
func InitDBConnection() error {
	// Get MySQL credentials from environment variables
	databaseHost := os.Getenv("DATABASE_HOST")
	databaseUser := os.Getenv("DATABASE_USER")
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_NAME")

	// Create the connection string for MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", databaseUser, databasePassword, databaseHost, databaseName)

	// Open a connection to the database
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	// Ping the database to check the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// Log successful connection
	log.Printf(`{"success": true, "message": "mysql database connected"}`)
	return nil
}
