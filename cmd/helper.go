package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func unauthorized(c *gin.Context, redirectTo string) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	} else {
		c.Redirect(http.StatusFound, redirectTo)
	}
}
