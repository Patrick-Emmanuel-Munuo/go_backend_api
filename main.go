package main

import (
	"log"
	"net/http"
	"runtime"
	"vartrick/controllers"
	"vartrick/helpers"
	"vartrick/route"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf(`{"success": false, "message": "No .env file found or failed to load"}`)
	}

	// Update helper vars from environment
	helpers.UpdateEnvVars()

	// Recover from panic (simulate try-catch)
	defer func() {
		if r := recover(); r != nil {
			log.Printf(`{"success": false, "message": "Unexpected fatal error", "error": "%v"}`, r)
		}
	}()

	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Gin release mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Trust only localhost (safe for dev)
	//router.SetTrustedProxies([]string{"127.0.0.1"})

	// Initialize DB once (no retries here)
	dbResult := helpers.InitDBConnection()
	if dbResult["success"].(bool) {
		controllers.SetDB(helpers.DB)
		log.Printf(`{"success": true, "message": "Database connected successfully"}`)
	} else {
		log.Printf(`{"success": false, "message": "%s"}`, dbResult["message"])
	}

	// Basic welcome route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Golang Welcome to VarTrick Server application",
		})
	})

	// Load app routes
	route.Router_main(router)
	route.Router_mysql(router)

	// Handle 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Wrong route or requested method",
		})
	})

	// Start server (no retry)
	result := helpers.StartServer(router)
	if !result["success"].(bool) {
		log.Printf(`{"success": false, "message": "%s"}`, result["message"])
	}
}
