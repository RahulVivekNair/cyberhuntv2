package main

import (
	"flag"
	"log"
	"net/http"

	"cyberhunt/internal/database"
	"cyberhunt/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	var addr = flag.String("addr", ":8080", "Address and port to run the server")
	flag.Parse()

	// Initialize database
	db, err := database.InitDB("postgres://game:secret@localhost:5432/cyberhunt?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Setup router
	r := gin.Default()

	// Load templates from embedded files
	r.LoadHTMLGlob("templates/*")

	// Static files

	// Routes

	// Public Routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/login", h.LoginPage)
	r.POST("/login", h.Login)

	// User Routes (require authentication)
	r.GET("/game", h.AuthMiddleware(), h.GamePage)
	r.POST("/api/scan", h.AuthMiddleware(), h.ScanQR)
	r.GET("/leaderboard", h.AuthMiddleware(), h.LeaderboardPage)

	// Admin Routes
	r.GET("/admin", h.AdminAuthMiddleware(), h.AdminPage)
	r.GET("/adminlogin", h.AdminLoginPage)
	r.POST("/adminlogin", h.AdminLogin)
	r.POST("/api/admin/start", h.AdminAuthMiddleware(), h.StartGame)
	r.POST("/api/admin/end", h.AdminAuthMiddleware(), h.EndGame)
	r.POST("/api/admin/clear", h.AdminAuthMiddleware(), h.ClearState)
	r.POST("/api/admin/group", h.AdminAuthMiddleware(), h.AddGroup)
	r.DELETE("/api/admin/group/:id", h.AdminAuthMiddleware(), h.DeleteGroup)
	r.GET("/api/admin/status", h.AdminAuthMiddleware(), h.GetGameStatus)
	r.GET("/api/admin/stats", h.AdminAuthMiddleware(), h.GetStats)
	r.GET("/api/admin/leaderboard", h.AdminAuthMiddleware(), h.AdminLeaderboard)
	r.POST("/logout", h.Logout)
	r.GET("/api/leaderboard", h.AuthMiddleware(), h.GetLeaderboard)

	// Seed routes
	r.GET("/seed", h.AdminAuthMiddleware(), h.SeedPage)
	r.POST("/api/seed/groups", h.AdminAuthMiddleware(), h.SeedGroups)
	r.POST("/api/seed/clues", h.AdminAuthMiddleware(), h.SeedClues)
	r.POST("/api/seed/total_clues", h.AdminAuthMiddleware(), h.UpdateTotalClues)

	// Start server
	log.Println("Server starting on", *addr)
	if err := r.Run(*addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
