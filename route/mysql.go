package route

import (
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
				c.JSON(http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}

			// Query the database for user
			result := controllers.Read(options)

			status := http.StatusInternalServerError
			if success, ok := result["success"].(bool); ok && success {
				status = http.StatusOK
			}

			// Check if user exists
			if success, ok := result["success"].(bool); ok && success {
				if messages, ok := result["message"].([]map[string]interface{}); ok && len(messages) > 0 {
					// Gather client metadata
					userBrowser := map[string]interface{}{
						"ip_address":      helpers.ClearIPAddress(c.ClientIP()),
						"host":            c.Request.Host,
						"os":              c.Request.UserAgent(), // can parse more precisely with a library
						"browser":         "",                    // optional: add browser parsing library if needed
						"browser_version": "",
					}
					// Merge metadata into first user record
					messages[0]["user_browser"] = userBrowser
					// Generate JWT token for the user
					authResult := helpers.Authenticate(map[string]interface{}{
						"id":       messages[0]["id"],
						"username": messages[0]["user_name"],
						"role":     messages[0]["role"],
					})
					if successToken, ok := authResult["success"].(bool); ok && successToken {
						// Attach token to user record
						messages[0]["token"] = authResult["message"]
					} else {
						c.JSON(http.StatusInternalServerError, map[string]interface{}{
							"success": false,
							"message": "Failed to generate token",
						})
						return
					}
					result["message"] = messages
				}
			}
			c.JSON(status, result)
		})

		mysql.POST("/read", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Read(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/read-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options []map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.ReadBulk(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/list", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.List(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/list-all", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.ListAll(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/update", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Update(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/update-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options []map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.UpdateBulk(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/create", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Create(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/create-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.CreateBulk(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/delete", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Delete(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/delete-bulk", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options []map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.DelateBulk(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/search", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Search(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/search-between", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.SearchBetween(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
		})
		mysql.POST("/count", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid JSON body",
				})
				return
			}
			response := controllers.Count(options)
			status := http.StatusInternalServerError
			if success, ok := response["success"].(bool); ok && success {
				status = http.StatusOK
			}
			c.JSON(status, response)
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
	}
}
