package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *Handler) Login(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	group, err := h.groupService.GetGroupByNameAndPassword(c.Request.Context(), name, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login!"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"groupID": group.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("auth", tokenString, 3600*24, "/", "", false, true)

	// Respond with success (JSON)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) AdminLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "adminLogin.html", nil)
}

func (h *Handler) AdminLogin(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	admin, err := h.adminService.GetAdminByNameAndPassword(c.Request.Context(), name, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login!"})
		return
	}

	// Create JWT token for admin
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"adminID": admin.ID,
		"isAdmin": true,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("adminAuth", tokenString, 3600*24, "/", "", false, true)

	// Respond with success instead of redirect
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", false, true)
	c.SetCookie("adminAuth", "", -1, "/", "", false, true)
}
