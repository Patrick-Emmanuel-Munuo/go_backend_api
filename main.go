package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
	"vartrick/controllers"
	"vartrick/helpers"
	"vartrick/route"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// ColorLogger prints colored, pretty logs for each HTTP request
func ColorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()

		// Colorize status codes
		var statusColor func(a ...interface{}) string
		switch {
		case status >= 200 && status < 300:
			statusColor = color.New(color.FgGreen).SprintFunc()
		case status >= 300 && status < 400:
			statusColor = color.New(color.FgCyan).SprintFunc()
		case status >= 400 && status < 500:
			statusColor = color.New(color.FgYellow).SprintFunc()
		default:
			statusColor = color.New(color.FgRed).SprintFunc()
		}

		methodColor := color.New(color.FgBlue, color.Bold).SprintFunc()
		pathColor := color.New(color.FgMagenta).SprintFunc()
		ipColor := color.New(color.FgHiWhite).SprintFunc()
		latencyColor := color.New(color.FgHiCyan).SprintFunc()

		fmt.Printf("%s | %s | %s | %s | %s\n",
			ipColor(c.ClientIP()),
			statusColor(status),
			methodColor(c.Request.Method),
			pathColor(c.Request.URL.Path),
			latencyColor(latency),
		)
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		helpers.LogJSON(false, fmt.Sprintf("error in Load environment variables Unexpected fatal : %v", err))
	}

	// Update helper vars
	helpers.UpdateEnvVars()

	// --- Recover from panic ---
	defer func() {
		if r := recover(); r != nil {
			helpers.LogJSON(false, fmt.Sprintf("error in Recover from panic Unexpected fatal error: %v", r))
		}
	}()

	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Gin setup
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(204) // No Content
	})
	// --- Middlewares ---
	router.Use(ColorLogger())  // pretty colored request logs
	router.Use(gin.Recovery()) // catch panics

	// Initialize DB
	dbResult := helpers.InitDBConnection()
	if dbResult["success"].(bool) {
		controllers.SetDB(helpers.DB)
		helpers.LogJSON(true, "Database connected successfully")
	} else {
		helpers.LogJSON(false, fmt.Sprintf("Database connection failed: %s", dbResult["message"]))
	}

	// Basic route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Golang Welcome to VarTrick Server application",
		})
	})

	// Start cleanup to prevent memory leaks
	helpers.StartCleanup(10 * time.Minute)

	// Rate limiter
	router.Use(helpers.RateLimitMiddleware(2, 3, 10*time.Second, "/api/", "/api/V1/"))

	// Load app routes
	route.Router_main(router)
	route.Router_mysql(router)

	// Handle 404
	router.NoRoute(func(c *gin.Context) {
		helpers.LogJSON(false, fmt.Sprintf("404 Not Found: %s %s", c.Request.Method, c.Request.URL.Path))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Wrong route or requested method",
		})
	})

	// Start server
	result := helpers.StartServer(router)
	if result["success"].(bool) {
		helpers.LogJSON(true, fmt.Sprintf("Server started successfully on port %s", helpers.ServerPort))
	} else {
		helpers.LogJSON(false, fmt.Sprintf("Server failed to start: %s", result["message"]))
	}
}
