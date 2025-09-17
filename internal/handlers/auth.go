package handlers

import (
	"cyberhunt/internal/models"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	db        *sql.DB
	jwtSecret string
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db:        db,
		jwtSecret: "05f3b711c3722735c25ddc7587cb9cb2", // Change this in production
	}
}

func (h *Handler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *Handler) Login(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	var group models.Group
	err := h.db.QueryRow(`
		SELECT id, name, pathway, current_clue_idx, completed, end_time, password
		FROM groups WHERE name = ? AND password = ?
	`, name, password).Scan(
		&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
		&group.Completed, &group.EndTime, &group.Password,
	)

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

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth")
		if err != nil {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("groupID", int(claims["groupID"].(float64)))
		c.Next()
	}
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", false, true)
	c.SetCookie("adminAuth", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) AdminLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "adminLogin.html", nil)
}

func (h *Handler) AdminLogin(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	var admin models.Admin
	err := h.db.QueryRow(`
		SELECT id, name, password FROM admins WHERE name = ? AND password = ?
	`, name, password).Scan(&admin.ID, &admin.Name, &admin.Password)

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

func (h *Handler) AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("adminAuth")
		if err != nil {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusFound, "/adminlogin")
			}
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusFound, "/adminlogin")
			}
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if isAdmin, ok := claims["isAdmin"].(bool); !ok || !isAdmin {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusFound, "/adminlogin")
			}
			c.Abort()
			return
		}

		c.Set("adminID", int(claims["adminID"].(float64)))
		c.Next()
	}
}
