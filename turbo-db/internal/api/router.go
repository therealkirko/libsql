package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourname/turbo-db/internal/config"
)

// SetupRouter sets up the API router
func SetupRouter(handler *Handler, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// CORS middleware (customize as needed)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication middleware (if enabled)
		if cfg.EnableAuth {
			v1.Use(AuthMiddleware(cfg.APIKey))
		}

		// Database routes
		databases := v1.Group("/databases")
		{
			databases.POST("", handler.CreateDatabase)
			databases.GET("", handler.ListDatabases)
			databases.GET("/:id", handler.GetDatabase)
			databases.DELETE("/:id", handler.DeleteDatabase)
			databases.POST("/:id/stop", handler.StopDatabase)
			databases.GET("/:id/connection", handler.GetDatabaseConnection)
		}
	}

	return router
}

// AuthMiddleware provides API key authentication
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.GetHeader("Authorization")
			// Remove "Bearer " prefix if present
			if len(key) > 7 && key[:7] == "Bearer " {
				key = key[7:]
			}
		}

		if key != apiKey {
			c.JSON(401, gin.H{
				"error": "unauthorized",
				"message": "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
