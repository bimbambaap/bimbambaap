package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bimbambaap/bimbambaap/handlers"
	"github.com/bimbambaap/bimbambaap/middleware"
)

func Setup() *gin.Engine {
	r := gin.Default()

	// Rate limiting per IP via een simpele middleware
	r.Use(rateLimiter())

	// CORS — pas aan naar jouw frontend URL in productie
	r.Use(corsMiddleware())

	// Health check voor Railway
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		// Auth routes — geen token nodig
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// User routes
		users := api.Group("/users")
		{
			users.GET("/:username/posts", handlers.GetUserPosts) // publiek
			users.Use(middleware.Auth())
			{
				users.GET("/me", handlers.GetMe)
				users.PUT("/me", handlers.UpdateProfile)
				users.PUT("/:id/admin", handlers.SetAdmin)
			}
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.Auth())
		{
			admin.GET("/users", handlers.AdminGetUsers)
			admin.DELETE("/users/:id", handlers.AdminDeleteUser)
			admin.GET("/posts", handlers.AdminGetPosts)
		}

		// Post routes
		posts := api.Group("/posts")
		{
			posts.GET("/feed", handlers.GetFeed)       // publiek
			posts.GET("/:id", handlers.GetPost)        // publiek
			posts.Use(middleware.Auth())
			{
				posts.POST("", handlers.CreatePost)
				posts.DELETE("/:id", handlers.DeletePost)
			}
		}
	}

	return r
}

// Simpele in-memory rate limiter
var requestCounts = make(map[string][]time.Time)

func rateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		window := time.Minute

		// Verwijder oude requests
		valid := []time.Time{}
		for _, t := range requestCounts[ip] {
			if now.Sub(t) < window {
				valid = append(valid, t)
			}
		}
		requestCounts[ip] = append(valid, now)

		// Max 60 requests per minuut
		if len(requestCounts[ip]) > 60 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Te veel requests, wacht even"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // Verander naar jouw domein in productie
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
