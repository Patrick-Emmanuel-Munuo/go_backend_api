package route

import (
	"net/http"
	"vartrick/controllers"
	"vartrick/helpers"

	"github.com/gin-gonic/gin"
)

// helper: decrypt and validate incoming payload
func decryptAndValidate(c *gin.Context) (map[string]interface{}, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
			"success": false,
			"message": "Invalid JSON body",
		}))
		return nil, false
	}

	decrypted := helpers.Decript(body)
	if success, ok := decrypted["success"].(bool); !ok || !success {
		c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
			"success": false,
			"message": "Failed to decrypt data. Check encryption keys.",
		}))
		return nil, false
	}

	message, ok := decrypted["message"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
			"success": false,
			"message": "Invalid decrypted payload",
		}))
		return nil, false
	}

	return message, true
}

// helper: determine HTTP status
func determineStatus(response map[string]interface{}) int {
	if success, ok := response["success"].(bool); ok && success {
		return http.StatusOK
	}
	return http.StatusInternalServerError
}

func Router_mysql(router *gin.Engine) {
	mysql := router.Group("/api/V1")
	{
		// Base route
		mysql.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Backend Mysql API router is working",
			})
		})

		// LOGIN
		mysql.POST("/login", func(c *gin.Context) {
			message, ok := decryptAndValidate(c)
			if !ok {
				return
			}

			response := controllers.Read(message)

			status := determineStatus(response)

			if messages, ok := response["message"].([]map[string]interface{}); ok && len(messages) > 0 {
				user := messages[0]
				user["user_browser"] = map[string]interface{}{
					"ip_address":      c.ClientIP(),
					"host":            c.Request.Host,
					"os":              c.Request.UserAgent(),
					"browser_version": c.Request.UserAgent(),
					"browser_name":    c.Request.UserAgent(),
				}

				authResult := helpers.Authenticate(map[string]interface{}{
					"id":        user["id"],
					"user_name": user["user_name"],
				})

				if successToken, ok := authResult["success"].(bool); ok && successToken {
					user["token"] = authResult["message"]
				} else {
					c.JSON(http.StatusInternalServerError, helpers.Encript(map[string]interface{}{
						"success": false,
						"message": "Failed to generate token",
					}))
					return
				}

				response["message"] = messages
			}

			c.JSON(status, helpers.Encript(response))
		})

		// SINGLE-OBJECT ROUTES
		singleRoutes := map[string]func(map[string]interface{}) map[string]interface{}{
			"/read":            controllers.Read,
			"/joint-read":      controllers.ReadJoin,
			"/list":            controllers.List,
			"/list-all":        controllers.ListAll,
			"/update":          controllers.Update,
			"/create":          controllers.Create,
			"/delete":          controllers.Delete,
			"/search":          controllers.Search,
			"/search-between":  controllers.SearchBetween,
			"/query":           controllers.Query,
			"/database-handle": controllers.DatabaseHandler,
		}

		for route, handler := range singleRoutes {
			r := route
			h := handler
			mysql.POST(r, helpers.AuthMiddleware(), func(c *gin.Context) {
				message, ok := decryptAndValidate(c)
				if !ok {
					return
				}
				response := h(message)
				c.JSON(determineStatus(response), helpers.Encript(response))
			})
		}

		// BULK ROUTES
		bulkRoutes := map[string]func([]map[string]interface{}) map[string]interface{}{
			"/read-bulk":   controllers.ReadBulk,
			"/create-bulk": func(msgs []map[string]interface{}) map[string]interface{} { return controllers.CreateBulk(msgs[0]) }, // adjust if controller expects array
			"/update-bulk": controllers.UpdateBulk,
			"/delete-bulk": controllers.DelateBulk,
			"/bulk-count":  controllers.CountBulk,
		}

		for route, handler := range bulkRoutes {
			r := route
			h := handler
			mysql.POST(r, helpers.AuthMiddleware(), func(c *gin.Context) {
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err != nil {
					c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
						"success": false,
						"message": "Invalid JSON body",
					}))
					return
				}

				options := helpers.Decript(body)
				if success, ok := options["success"].(bool); !ok || !success {
					c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
						"success": false,
						"message": "Failed to decrypt data. Check encryption keys.",
					}))
					return
				}

				messageRaw := options["message"]
				var message []map[string]interface{}

				switch v := messageRaw.(type) {
				case []interface{}:
					for _, item := range v {
						if m, ok := item.(map[string]interface{}); ok {
							message = append(message, m)
						} else {
							c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
								"success": false,
								"message": "Invalid array element type",
							}))
							return
						}
					}
				case map[string]interface{}:
					message = append(message, v)
				default:
					c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
						"success": false,
						"message": "Invalid decrypted payload type",
					}))
					return
				}

				response := h(message)
				c.JSON(determineStatus(response), helpers.Encript(response))
			})
		}

		// SINGLE ROUTE: BACKUP
		mysql.POST("/backup", helpers.AuthMiddleware(), func(c *gin.Context) {
			var options map[string]interface{}
			if err := c.ShouldBindJSON(&options); err != nil {
				c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
					"success": false,
					"message": "Invalid JSON body",
				}))
				return
			}
			response := controllers.Backup(options)
			c.JSON(determineStatus(response), helpers.Encript(response))
		})

		// COUNT ROUTE: supports single and bulk
		mysql.POST("/count", helpers.AuthMiddleware(), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
					"success": false,
					"message": "Invalid JSON body",
				}))
				return
			}

			options := helpers.Decript(body)
			if success, ok := options["success"].(bool); !ok || !success {
				c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
					"success": false,
					"message": "Failed to decrypt data. Check encryption keys.",
				}))
				return
			}

			messageRaw := options["message"]
			var response map[string]interface{}

			switch v := messageRaw.(type) {
			case map[string]interface{}:
				response = controllers.Count(v)
			case []interface{}:
				var bulk []map[string]interface{}
				for _, item := range v {
					if m, ok := item.(map[string]interface{}); ok {
						bulk = append(bulk, m)
					} else {
						c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
							"success": false,
							"message": "Invalid array element type",
						}))
						return
					}
				}
				response = controllers.CountBulk(bulk)
			default:
				c.JSON(http.StatusBadRequest, helpers.Encript(map[string]interface{}{
					"success": false,
					"message": "Invalid decrypted payload type",
				}))
				return
			}

			c.JSON(determineStatus(response), helpers.Encript(response))
		})
	}
}
