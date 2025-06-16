package route

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"vartrick/controllers" // Assuming you'll have models for MySQL connection

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	routes := router.Group("/api")
	{
		// Base route for testing if the server is running
		routes.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Backend MySQL API application router is working",
			})
		})
		//get server time /server/time
		routes.GET("/server/time", func(c *gin.Context) {
			now := time.Now()
			zone, offset := now.Zone()
			message := map[string]interface{}{
				"local_time":      now.Format("2006-01-02 15:04:05"),       // local time
				"utc_time":        now.UTC().Format("2006-01-02 15:04:05"), // UTC time
				"iso8601_time":    now.Format(time.RFC3339),                // ISO 8601 format
				"timezone":        zone,
				"timezone_offset": offset,     // in seconds
				"unix_timestamp":  now.Unix(), // UNIX timestamp
				"status":          "server time retrieved successfully",
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": message,
			})
		})
		//get client data,ip adress,....etc
		routes.GET("/client", func(c *gin.Context) {
			clientIP := c.ClientIP()
			userAgent := c.Request.UserAgent()
			// You can also grab other headers if needed
			acceptLang := c.GetHeader("Accept-Language")
			referer := c.GetHeader("Referer")

			message := map[string]interface{}{
				"ip_address":     clientIP,
				"user_agent":     userAgent,
				"accept_lang":    acceptLang,
				"referer":        referer,
				"request_path":   c.Request.URL.Path,
				"request_method": c.Request.Method,
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": message,
			})
		})

		//this router encript -token
		routes.GET("/encript-token", func(c *gin.Context) {
			amount := c.Query("amount")
			result := controllers.EncriptToken(map[string]interface{}{
				"amount": amount,
			})
			if success, ok := result["success"].(bool); ok && success {
				c.JSON(http.StatusOK, result)
			} else {
				c.JSON(http.StatusInternalServerError, result)
			}
		})
		// This route decript-token
		routes.GET("/decript-token", func(c *gin.Context) {
			token := c.Query("token")
			result := controllers.DecriptToken(map[string]interface{}{
				"token": token,
			})
			if success, ok := result["success"].(bool); ok && success {
				c.JSON(http.StatusOK, result)
			} else {
				c.JSON(http.StatusInternalServerError, result)
			}
		})
		// This route generates a one-time password (OTP) for testing purposes
		routes.GET("/generate-otp", func(c *gin.Context) {
			result := controllers.GenerateOTP()
			if success, ok := result["success"].(bool); ok && success {
				c.JSON(http.StatusOK, result)
			} else {
				c.JSON(http.StatusInternalServerError, result)
			}
		})
		//This route generates and send otp one-time password (OTP) for testing purposes
		routes.GET("/send-otp", func(c *gin.Context) {
			result := controllers.GenerateOTP()
			if success, ok := result["success"].(bool); ok && success {
				c.JSON(http.StatusOK, result)
			} else {
				c.JSON(http.StatusInternalServerError, result)
			}
		})
		//send sms routers
		routes.GET("/send-sms-local", func(c *gin.Context) {
			to := c.Query("to")
			message := c.Query("message")
			if to == "" || message == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "'to' and 'message' query parameters are required"})
				return
			}
			responce := controllers.SendMessageLocal(map[string]interface{}{
				"to":      strings.Split(to, ","),
				"message": message,
			})
			if success, ok := responce["success"].(bool); ok && success {
				c.JSON(http.StatusOK, responce)
			} else {
				c.JSON(http.StatusInternalServerError, responce)
			}
		})
		//send sms routers
		routes.GET("/send-sms", func(c *gin.Context) {
			to := c.Query("to")
			message := c.Query("message")
			if to == "" || message == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "'to' and 'message' query parameters are required"})
				return
			}
			responce := controllers.SendMessage(map[string]interface{}{
				"to":      strings.Split(to, ","),
				"message": message,
			})
			if success, ok := responce["success"].(bool); ok && success {
				c.JSON(http.StatusOK, responce)
			} else {
				c.JSON(http.StatusInternalServerError, responce)
			}
		})
		//send mail routers
		routes.GET("/send-mail", func(c *gin.Context) {
			to := c.Query("to") // e.g. "user1@example.com,user2@example.com"
			message := c.Query("message")
			if to == "" || message == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "'to' and 'message' query parameters are required"})
				return
			}
			// Split multiple emails by comma, trim spaces
			recipients := strings.Split(to, ",")
			for i := range recipients {
				recipients[i] = strings.TrimSpace(recipients[i])
			}
			options := map[string]interface{}{
				"To":      recipients,
				"Message": message,
				"Subject": "Test Email",
				"HTML":    "<b>This is bold</b>",
				// "Attachments": []string{"./report.pdf"},
			}
			responce := controllers.SendMail(options)
			if success, ok := responce["success"].(bool); ok && success {
				c.JSON(http.StatusOK, responce)
			} else {
				c.JSON(http.StatusInternalServerError, responce)
			}
		})
	}
	// Example of MySQL routes (commented for now)
	mysql := router.Group("/api/mysql")
	{
		mysql.POST("/read", func(c *gin.Context) {
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
		mysql.POST("/read-bulk", func(c *gin.Context) {
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
		mysql.POST("/list", func(c *gin.Context) {
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
		mysql.POST("/list-all", func(c *gin.Context) {
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
		mysql.POST("/update", func(c *gin.Context) {
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
		mysql.POST("/update-bulk", func(c *gin.Context) {
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
		mysql.POST("/create", func(c *gin.Context) {
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
		mysql.POST("/create-bulk", func(c *gin.Context) {
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
		mysql.POST("/delete", func(c *gin.Context) {
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
		mysql.POST("/delete-bulk", func(c *gin.Context) {
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
		mysql.POST("/search", func(c *gin.Context) {
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
		mysql.POST("/search-between", func(c *gin.Context) {
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
		mysql.POST("/count", func(c *gin.Context) {
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
		mysql.POST("/backup", func(c *gin.Context) {
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
	files := router.Group("/api/files")
	{
		//files.POST("/upload", controllers.UploadFileHandler)
		//files.POST("/upload-multiple", controllers.UploadMultipleFilesHandler)
		//files.GET("/download/:filename", controllers.DownloadFile)
		files.GET("/download", func(c *gin.Context) {
			filename := c.Query("file")
			if filename == "" {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "file name required in query",
				})
				return
			}
			fullPath := filepath.Join("./public", filename)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "file name not found",
				})
				return
			}
			// Instead of returning map here, serve the file directly
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
			c.Header("Content-Type", "application/octet-stream")
			c.File(fullPath)
			return // no JSON response because we sent file directly
		})
		//files.DELETE("/delete/:filename", controllers.DeleteFileHandler)
	}

}
