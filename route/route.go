package route

import (
	"net/http"
	"vartrick-server/controllers" // Assuming you'll have models for MySQL connection

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	// Connect to your database here
	// db := models.ConnectDatabase()
	Routes := router.Group("/api")
	{
		// Base route for testing if the server is running
		Routes.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Backend MySQL API application router is working",
			})
		})

		// Generate OTP route
		Routes.GET("/generate-otp", func(c *gin.Context) {
			otp := controllers.GenerateOTP() // Generate OTP
			if otp == "" {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to generate OTP",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": otp,
				})
			}
		})
	}

	// Example of MySQL routes (commented for now)
	// mysql := router.Group("/api/mysql")
	// {
	// 	mysql.POST("/read", ReadData)         // Your controller logic here
	// 	mysql.POST("/bulk-read", BulkReadData)
	// 	mysql.POST("/list", ListData)
	// 	mysql.GET("/list-all", ListAllData)
	// 	mysql.POST("/create", CreateData)
	// 	mysql.POST("/bulk-create", BulkCreateData)
	// 	mysql.POST("/update", UpdateData)
	// 	mysql.POST("/bulk-update", BulkUpdateData)
	// 	mysql.POST("/delete", DeleteData)
	// 	mysql.POST("/bulk-delete", BulkDeleteData)
	// 	mysql.POST("/authentication", Authentication)
	// 	mysql.POST("/count", CountRecords)
	// 	mysql.POST("/search", SearchData)
	// 	mysql.GET("/backup", BackupDatabase)
	// 	mysql.POST("/create-database", CreateDatabase)
	// 	mysql.POST("/delete-database", DeleteDatabase)
	// 	mysql.POST("/create-table", CreateTable)
	// 	mysql.POST("/delete-table", DeleteTable)
	// 	mysql.POST("/update-table", UpdateTable)
	// }
}
