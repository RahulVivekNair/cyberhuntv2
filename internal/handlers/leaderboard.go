package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LeaderboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "leaderboard.html", nil)
}
