package route

import (
	"fmt"
	"net/http"
	"vartrick/controllers" // Assuming you'll have models for MySQL connection
	"vartrick/helpers"

	"github.com/gin-gonic/gin"
)

func Router_mysql(router *gin.Engine) {
	// Example of MySQL routes (commented for now)
	mysql := router.Group("/api/v1")
	{
		// Base route for testing if the server is running
		mysql.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Backend Mysql Api application router is working",
			})
		})
		mysql.POST("/login", func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			// Decrypt incoming data
			decrypted := helpers.Decript(options)
			if success, ok := decrypted["success"].(bool); !ok || !success {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Failed to decrypt data. Check encryption keys.",
				})
				return
			}

			message, ok := decrypted["message"].(map[string]interface{})
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid decrypted payload",
				})
				return
			}

			// Query user from database
			response := controllers.Read(message)

			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}

			// If user found, attach metadata and generate JWT
			if messages, ok := response["message"].([]map[string]interface{}); ok && len(messages) > 0 {
				user := messages[0]
				user["user_browser"] = map[string]interface{}{
					"ip_address": c.ClientIP(),
					"host":       c.Request.Host,
					"os":         c.Request.UserAgent(),
				}

				authResult := helpers.Authenticate(map[string]interface{}{
					"id":        user["id"],
					"user_name": user["user_name"],
				})

				if successToken, ok := authResult["success"].(bool); ok && successToken {
					user["token"] = authResult["message"]
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"message": "Failed to generate token",
					})
					return
				}

				response["message"] = messages
			}
			fmt.Println(response)
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})

		mysql.POST("/read", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			// Query user from database
			response := controllers.Read(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, response)
		})
		mysql.POST("/joint-read", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.ReadJoin(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/read-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body []map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.ReadBulk(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/list", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.List(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/list-all", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.ListAll(body)
			// Encrypt response before sending
			encryptedResponse := helpers.Encript(response)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, encryptedResponse)
		})
		mysql.POST("/update", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.Update(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/update-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body []map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.UpdateBulk(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/create", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.Create(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			fmt.Println()
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/create-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.CreateBulk(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/delete", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.Delete(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/delete-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body []map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.DelateBulk(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/search", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.Search(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/search-between", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.SearchBetween(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/count", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.Count(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/count-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body []map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.CountBulk(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/backup", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Backup(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/query", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.Query(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
		mysql.POST("/database-handle", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			response := controllers.DatabaseHandler(body)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			// Encrypt the response if encryption is enabled
			c.JSON(status, helpers.Encript(response))
		})
	}
}
