package main

import (
	"net/http"
	"strings"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	JWTSecret string
}

// AuthMiddleware protects regular users
func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth")
		if err != nil {
			unauthorized(c, "/login")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			unauthorized(c, "/login")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			groupIDFloat, ok := claims["groupID"].(float64)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}
			c.Set("groupID", int(groupIDFloat))
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminAuthMiddleware protects admin users
func (m *Middleware) AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("adminAuth")
		if err != nil {
			unauthorized(c, "/adminlogin")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.JWTSecret), nil
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

		// Validate claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check if user is admin
			isAdmin, ok := claims["isAdmin"].(bool)
			if !ok || !isAdmin {
				unauthorized(c, "/adminlogin")
				c.Abort()
				return
			}

			// Set adminID in context
			adminIDFloat, ok := claims["adminID"].(float64)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}
			c.Set("adminID", int(adminIDFloat))
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}
