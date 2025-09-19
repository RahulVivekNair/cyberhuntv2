package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LeaderboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "leaderboard.html", nil)
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	// Get total_clues
	totalClues, _ := h.gameService.GetTotalClues(c.Request.Context())

	settings, _ := h.gameService.GetGameStatus(c.Request.Context())

	groupsFromDB, err := h.groupService.GetGroupsForLeaderboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	var groups []gin.H
	rank := 1
	for _, group := range groupsFromDB {
		// Calculate total time if completed and game actually started
		var totalTime string
		if group.Completed && group.EndTime != nil && settings.StartTime != nil {
			duration := group.EndTime.Sub(*settings.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			seconds := int(duration.Seconds()) % 60

			if hours > 0 {
				totalTime = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
			} else {
				totalTime = fmt.Sprintf("%dm %ds", minutes, seconds)
			}
		}

		// Add rank badge
		var badge string
		switch rank {
		case 1:
			badge = "🥇"
		case 2:
			badge = "🥈"
		case 3:
			badge = "🥉"
		}

		groups = append(groups, gin.H{
			"rank":             rank,
			"badge":            badge,
			"name":             group.Name,
			"pathway":          group.Pathway,
			"current_clue_idx": group.CurrentClueIdx,
			"completed":        group.Completed,
			"total_time":       totalTime,
		})

		rank++
	}
	c.JSON(http.StatusOK, gin.H{
		"groups":     groups,
		"totalClues": totalClues,
	})
}
