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

	group, err := h.groupService.GetGroupByNameAndPassword(name, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Invalid login!"})
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"groupID": group.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Failed to create session"})
		return
	}

	// Set cookie
	c.SetCookie("auth", tokenString, 3600*24, "/", "", false, true)
	c.Redirect(http.StatusFound, "/game")
}

func (h *Handler) AdminLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "adminLogin.html", nil)
}

func (h *Handler) AdminLogin(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	admin, err := h.adminService.GetAdminByNameAndPassword(name, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "adminLogin.html", gin.H{"error": "Invalid login!"})
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
		c.HTML(http.StatusInternalServerError, "adminLogin.html", gin.H{"error": "Failed to create session"})
		return
	}

	// Set cookie
	c.SetCookie("adminAuth", tokenString, 3600*24, "/", "", false, true)
	c.Redirect(http.StatusFound, "/admin")
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", false, true)
	c.SetCookie("adminAuth", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}
