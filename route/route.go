package route

import (
	"net/http"
	"strings"
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
		routes.GET("/send-sms", func(c *gin.Context) {
			to := "255625449295"  //c.Query("to")
			message := "fine pat" //c.Query("message")
			if to == "" || message == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "'to' and 'message' query parameters are required"})
				return
			}
			options := controllers.SMSOptions{
				To:      strings.Split(to, ","),
				Message: message,
			}
			responce := controllers.SendMessage(options)
			if success, ok := responce["success"].(bool); ok && success {
				c.JSON(http.StatusOK, responce)
			} else {
				c.JSON(http.StatusInternalServerError, responce)
			}
		})

		//send mail routers
		routes.GET("/send-mail", func(c *gin.Context) {
			to := "patrickmunuo98@gmail.com" //c.Query("to") // e.g. "user1@example.com,user2@example.com"
			message := "fine pat"            //c.Query("message")
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
		mysql.GET("/backup", func(c *gin.Context) {
			backup, err := controllers.Backup()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, backup)
		})
		mysql.POST("/read", controllers.Read)
		mysql.POST("/search", controllers.Search)
		mysql.POST("/create", controllers.Create)
		//mysql.POST("/delete", controllers.Delete) // Uncomment if you want to enable delete functionality
		mysql.POST("/update", controllers.Update)
		// New pagination / listing routes
		mysql.POST("/list", controllers.List)
		mysql.POST("/list-all", controllers.ListAll)

		// New bulk operation routes
		mysql.POST("/create-bulk", controllers.CreateBulk)
		mysql.POST("/update-bulk", controllers.UpdateBulk)

		// Other utility routes
		mysql.POST("/count", controllers.Count)
		mysql.POST("/query", controllers.Query)
		mysql.POST("/search-between", controllers.SearchBetween)
	}
}
