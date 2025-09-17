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
	db, err := database.InitDB("data/cyberhunt.db")
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
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/login", h.LoginPage)
	r.POST("/login", h.Login)
	r.GET("/game", h.AuthMiddleware(), h.GamePage)
	r.POST("/api/scan", h.AuthMiddleware(), h.ScanQR)
	r.GET("/leaderboard", h.LeaderboardPage)
	r.GET("/admin", h.AdminAuthMiddleware(), h.AdminPage)
	r.GET("/adminlogin", h.AdminLoginPage)
	r.POST("/adminlogin", h.AdminLogin)
	r.POST("/api/admin/start", h.AdminAuthMiddleware(), h.StartGame)
	r.POST("/api/admin/end", h.AdminAuthMiddleware(), h.EndGame)
	r.POST("/api/admin/clear", h.AdminAuthMiddleware(), h.ClearState)
	r.POST("/api/admin/group", h.AdminAuthMiddleware(), h.AddGroup)
	r.DELETE("/api/admin/group/:id", h.AdminAuthMiddleware(), h.DeleteGroup)
	r.GET("/api/admin/stats", h.AdminAuthMiddleware(), h.GetStats)
	r.GET("/api/admin/leaderboard", h.AdminAuthMiddleware(), h.AdminLeaderboard)
	r.GET("/api/leaderboard", h.GetLeaderboard) // Public leaderboard API
	r.POST("/logout", h.Logout)

	// Start server
	log.Println("Server starting on", *addr)
	if err := r.Run(*addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
