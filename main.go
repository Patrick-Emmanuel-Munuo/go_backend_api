package main

import (
	"database/sql"

	"fmt"
	"log"
	"net/http"
	"os"
	"vartrick-server/helpers"
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
	if err := godotenv.Load(); err != nil {
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
	if err := InitDBConnection(); err != nil {
		log.Printf(`{"success": false, "message": "Failed to connect to MySQL", "error": "%v"}`, err)
		os.Exit(1)
	} else {
		log.Printf(`{"success": true, "message": "Connected to MySQL successfully"}`)
	}
	// Basic welcome route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Golang Welcome to VarTrick Server application",
		})
	})
	// Use POST for Read because you expect JSON body
	router.POST("/read", Read)
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

func Read(c *gin.Context) {
	//test options Options
	// Hardcoded test input
	/*options := Options{
		Table:  "users",
		Select: []string{"id", "name", "email"},
		Condition: map[string]interface{}{
			"status": "active",
		},
		OrCondition: map[string]interface{}{
			"role": "admin",
		},
	}*/

	// Bind JSON input to Options struct
	var options Options

	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, ReadResult{Success: false, Message: "Invalid request format"})
		return
	}

	if options.Table == "" || (len(options.Condition) == 0 && len(options.OrCondition) == 0) {
		c.JSON(http.StatusBadRequest, ReadResult{Success: false, Message: "Missing table name or condition(s)"})
		return
	}

	selectFields := "*"
	if len(options.Select) > 0 {
		selectFields = helpers.JoinFields(options.Select)
	} else {
		selectFields = "*"
	}

	whereClause := ""
	if len(options.Condition) > 0 && len(options.OrCondition) > 0 {
		whereClause = fmt.Sprintf("( %s ) AND ( %s )",
			helpers.Where(options.Condition),
			helpers.WhereOr(options.OrCondition))
	} else if len(options.Condition) > 0 {
		whereClause = helpers.Where(options.Condition)
	} else if len(options.OrCondition) > 0 {
		whereClause = helpers.WhereOr(options.OrCondition)
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectFields, options.Table, whereClause)
	rows, err := db.Query(query)
	//fmt.Println("queery : ", query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReadResult{Success: false, Message: err.Error()})
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReadResult{Success: false, Message: err.Error()})
		return
	}
	var results []map[string]interface{}

	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))

		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			c.JSON(http.StatusInternalServerError, ReadResult{Success: false, Message: err.Error()})
			return
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[col] = *val
		}

		results = append(results, rowMap)
	}
	// Convert []uint8 to string
	for i, row := range results {
		newRow := make(map[string]interface{})
		for k, v := range row {
			if byteVal, ok := v.([]uint8); ok {
				newRow[k] = string(byteVal)
			} else {
				newRow[k] = v
			}
		}
		results[i] = newRow
	}
	if len(results) == 0 {
		c.JSON(http.StatusOK, ReadResult{
			Success: false,
			Message: "No data found",
		})
		return
	}
	c.JSON(http.StatusOK, ReadResult{
		Success: true,
		Message: results,
	})
}
