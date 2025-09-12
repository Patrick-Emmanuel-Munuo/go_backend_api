package main

import (
	"log"
	"net/http"
	"runtime"
	"time"
	"vartrick/controllers"
	"vartrick/helpers"
	"vartrick/route"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Printf(`{"success": false, "message": "No .env file found or failed to load"}`)
	}

	// Update helper vars from environment
	helpers.UpdateEnvVars()
	// --- Initialize Logger ---
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Info("Starting VarTrick Server...")

	// --- Recover from panic ---
	defer func() {
		if r := recover(); r != nil {
			sugar.Errorf("Unexpected fatal error: %v", r)
		}
	}()

	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Gin release mode
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery()) // catch panics and log
	// --- Logging Middleware ---
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		sugar.Infof("%s | %d | %s | %s", c.ClientIP(), status, c.Request.Method, c.Request.URL.Path)
		sugar.Infof("Latency: %v", latency)
	})
	// Apply the headers middleware globally
	//router.Use(middlewares.HeadersMiddleware())

	// Trust only localhost (safe for dev)
	//router.SetTrustedProxies([]string{"127.0.0.1"})

	// Initialize DB once (no retries here)
	dbResult := helpers.InitDBConnection()
	if dbResult["success"].(bool) {
		controllers.SetDB(helpers.DB)
		sugar.Info("Database connected successfully")
	} else {
		sugar.Errorf("Database connection failed: %s", dbResult["message"])
	}

	// Basic welcome route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Golang Welcome to VarTrick Server application",
		})
	})
	// Start cleanup to prevent memory leaks
	helpers.StartCleanup(10 * time.Minute)
	// ApiRateLimiter creates a rate limiter for all requests
	// 5 requests/sec per IP max burst 10  block 10 seconds for /api/ and /api/V1/ routes
	router.Use(helpers.RateLimitMiddleware(2, 3, 10*time.Second, "/api/", "/api/V1/"))

	// CORS middleware
	//router.Use(helpers.CORSMiddleware([]string{"http://localhost:3000"}))
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

	// Start server with logger
	result := helpers.StartServer(router, sugar)
	if result["success"].(bool) {
		sugar.Infof("Server started successfully: %s", result["message"])
	} else {
		sugar.Errorf("Server failed to start: %s", result["message"])
	}
}
