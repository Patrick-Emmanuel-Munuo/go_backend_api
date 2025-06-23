package main

import (
	"database/sql"
	"runtime"

	"fmt"
	"log"
	"net/http"
	"os"
	"vartrick/controllers"
	"vartrick/helpers"
	"vartrick/route"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// This assumes a global DB variable
var db *sql.DB

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf(`{"success": false, "message": "No .env file found or failed to load"}`)
	}
	// Now environment variables are loaded, update the helpers vars if needed
	helpers.UpdateEnvVars()
	// Recover from panic (simulating try-catch)
	defer func() {
		if r := recover(); r != nil {
			log.Printf(`{"success": false, "message": "Unexpected fatal error", "error": "%v"}`, r)
			os.Exit(1)
		}
	}()
	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Construct the full server URL hide port if(helpers.ServerSecurity == "https") and !helpers.SslCertificate == "" || !helpers.SslKey == ""

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// Initialize and check the DB connection with JSON logging Initialize DB connection
	result := InitDBConnection()
	if !result["success"].(bool) {
		log.Printf(`{"success": false, "message": "%s"}`, result["message"])
	} else {
		log.Printf(`{"success": true, "message": "Connected to MySQL successfully"}`)
	}
	// Inject DB into controller package
	controllers.SetDB(db)
	/* Start SMS reader in background
	go func() {
		for {
			response := controllers.ReadMessageLocal()
			if success, ok := response["success"].(bool); ok && success {
				log.Println(response["message"])
				//log.Printf(`{"success": true, "message": "Connected to MySQL successfully"}`)
			} else {
				//fmt.Println("SMS read error:", response["message"])
			}
			time.Sleep(1 * time.Second) // Adjust polling interval
		}
	}()*/

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
			"message": "Wrong route or requested method",
		})
	})
	// Start the server with HTTP or HTTPS based on environment variable
	var serverErr error
	if helpers.ServerSecurity == "https" {
		if helpers.SslCertificate == "" || helpers.SslKey == "" {
			log.Printf(`{"success": false, "message": "SSL Certificate or SSL Key environment variables are not set"}`)
			Url := fmt.Sprintf("http://%s:%s", helpers.GetServerIPAddress(), helpers.ServerPort)
			log.Printf(`{"success": true, "message": "Starting Gin HTTP server on %s"}`, Url)
			serverErr = router.Run("0.0.0.0:" + helpers.ServerPort)
		} else {
			Url := fmt.Sprintf("https://%s", helpers.GetServerIPAddress())
			log.Printf(`{"success": true, "message": "Starting Gin HTTPS server on %s"}`, Url)
			serverErr = router.RunTLS("0.0.0.0:443", helpers.SslCertificate, helpers.SslKey)

			if serverErr != nil {
				log.Printf(`{"success": false, "message": "Failed to start HTTPS server, fallback to HTTP", "error": "%v"}`, serverErr)

				// Log fallback as HTTP correctly
				fallbackURL := fmt.Sprintf("http://%s:%s", helpers.GetServerIPAddress(), helpers.ServerPort)
				log.Printf(`{"success": true, "message": "Starting Gin HTTP server on %s"}`, fallbackURL)

				serverErr = router.Run("0.0.0.0:" + helpers.ServerPort)
			}
		}
	} else {
		// Log fallback as HTTP correctly
		fallbackURL := fmt.Sprintf("http://%s:%s", helpers.GetServerIPAddress(), helpers.ServerPort)
		log.Printf(`{"success": true, "message": "Starting Gin HTTP server on %s"}`, fallbackURL)
		serverErr = router.Run("0.0.0.0:" + helpers.ServerPort)
	}

	if serverErr != nil {
		log.Printf(`{"success": false, "message": "Failed to start Gin server", "error": "%v"}`, serverErr)
		os.Exit(1)
	}
}

// InitDBConnection establishes and checks the MySQL connection
func InitDBConnection() map[string]interface{} {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		helpers.DatabaseUser,
		helpers.DatabasePassword,
		helpers.DatabaseHost,
		helpers.DatabasePort,
		helpers.DatabaseName,
	)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to connect to MySQL: %v", err),
		}
	}
	if err = db.Ping(); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to ping MySQL: %v", err),
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": "MySQL connection established successfully",
	}
}
