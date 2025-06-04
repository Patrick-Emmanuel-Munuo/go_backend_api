package route

import (
	"net/http"
	"vartrick-server/controllers" // Assuming you'll have models for MySQL connection

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	// Connect to your database here
	// db := models.ConnectDatabase()
	routes := router.Group("/api")
	{
		// Base route for testing if the server is running
		routes.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Backend MySQL API application router is working",
			})
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
		//mysql.POST("/read", )
		//mysql.POST("/list", ListData)
		//mysql.POST("/create", CreateData)
		//mysql.POST("/update", UpdateData)

	}
}
