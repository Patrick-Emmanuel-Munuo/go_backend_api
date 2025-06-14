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
	//url := fmt.Sprintf("%s://%s:%s", helpers.ServerSecurity, helpers.GetServerIPAddress(), helpers.ServerPort)
	var server_url string
	if helpers.ServerSecurity == "https" && helpers.SslCertificate != "" && helpers.SslKey != "" {
		// Default HTTPS port 443 - hide port in URL
		server_url = fmt.Sprintf("%s://%s", helpers.ServerSecurity, helpers.GetServerIPAddress())
	} else {
		// Show port for other cases
		server_url = fmt.Sprintf("%s://%s:%s", helpers.ServerSecurity, helpers.GetServerIPAddress(), helpers.ServerPort)
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Initialize and check the DB connection with JSON logging Initialize DB connection
	if err := InitDBConnection(); err != nil {
		log.Printf(`{"success": false, "message": "Failed to connect to MySQL", "error": "%v"}`, err)
		//os.Exit(1)
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
			"message": "Wrong route or requested method",
		})
	})

	// Start the server with HTTP or HTTPS based on environment variable
	var server_err error
	if helpers.ServerSecurity == "https" {
		if helpers.SslCertificate == "" || helpers.SslKey == "" {
			log.Printf(`{"success": false, "message": "Ssl Certificate or Ssl Key environment variables are not set"}`)
			log.Printf(`{"success": true, "message": "Starting Gin HTTP server on %s"}`, server_url)
			server_err = router.Run(helpers.GetServerIPAddress() + ":" + helpers.ServerPort)
		} else {
			log.Printf(`{"success": true, "message": "Starting Gin HTTPS server on %s"}`, server_url)
			server_err = router.RunTLS(helpers.GetServerIPAddress()+":443", helpers.SslCertificate, helpers.SslKey)
		}
	} else {
		log.Printf(`{"success": true, "message": "Starting Gin HTTP server on %s"}`, server_url)
		server_err = router.Run(helpers.GetServerIPAddress() + ":" + helpers.ServerPort)
	}
	if server_err != nil {
		log.Printf(`{"success": false, "message": "Failed to start Gin server", "error": "%v"}`, server_err)
		os.Exit(1)
	}
}

// InitDBConnection establishes and checks the MySQL connection
func InitDBConnection() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		helpers.DatabaseUser, helpers.DatabasePassword, helpers.DatabaseHost, helpers.DatabaseName)

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
