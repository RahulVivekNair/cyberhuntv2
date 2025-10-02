package main

import (
	"cyberhunt/internal/handlers"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(h *handlers.Handler, jwtSecret string) *gin.Engine {
	// Setup router
	r := gin.Default()

	//use gzip encoding
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	// Load templates from embedded files
	r.LoadHTMLGlob("templates/*")

	// Static files
	r.Static("/static", "static/")
	// Routes

	//init authjwt
	m := Middleware{JWTSecret: jwtSecret}
	// Public Routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	//Public User and Admin Routes
	r.GET("/login", h.LoginPage)
	r.POST("/login", h.Login)
	r.GET("/admin/login", h.AdminLoginPage)
	r.POST("/admin/login", h.AdminLogin)
	r.POST("/logout", h.Logout)

	// User Routes (require authentication)
	r.GET("/game", m.AuthMiddleware(), h.GamePage)
	r.GET("/leaderboard", m.AuthMiddleware(), h.LeaderboardPage)
	r.GET("/api/leaderboard/stream", m.AuthMiddleware(), h.LeaderboardStream)
	r.POST("/api/scan", m.AuthMiddleware(), h.ScanQR)
	r.GET("/api/game-partial", m.AuthMiddleware(), h.GamePartial)

	// Admin Routes
	r.GET("/admin", m.AdminAuthMiddleware(), h.AdminPage)
	r.POST("/api/admin/start", m.AdminAuthMiddleware(), h.StartGame)
	r.POST("/api/admin/end", m.AdminAuthMiddleware(), h.EndGame)
	r.POST("/api/admin/clear", m.AdminAuthMiddleware(), h.ClearState)
	r.POST("/api/admin/group", m.AdminAuthMiddleware(), h.AddGroup)
	r.DELETE("/api/admin/group/:id", m.AdminAuthMiddleware(), h.DeleteGroup)
	r.GET("/api/admin/status", m.AdminAuthMiddleware(), h.GetGameStatus)
	r.GET("/api/admin/leaderboard/stream", m.AdminAuthMiddleware(), h.LeaderboardStream)

	// Seed routes
	r.GET("/seed", m.AdminAuthMiddleware(), h.SeedPage)
	r.POST("/api/seed/groups", m.AdminAuthMiddleware(), h.SeedGroups)
	r.POST("/api/seed/clues", m.AdminAuthMiddleware(), h.SeedClues)
	r.POST("/api/seed/total_clues", m.AdminAuthMiddleware(), h.UpdateTotalClues)

	// Clue management routes
	r.GET("/api/clues", m.AdminAuthMiddleware(), h.GetAllClues)
	r.POST("/api/clues", m.AdminAuthMiddleware(), h.AddClue)
	r.PUT("/api/clues/:id", m.AdminAuthMiddleware(), h.UpdateClue)
	r.DELETE("/api/clues/:id", m.AdminAuthMiddleware(), h.DeleteClue)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	return r
}
