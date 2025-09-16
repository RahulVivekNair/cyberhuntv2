package handlers

import (
	"cyberhunt/internal/database"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) LeaderboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "leaderboard.html", nil)
}

func (h *Handler) GetLeaderboard(c *gin.Context) {
	// Get game settings
	var totalClues int
	var startTime time.Time
	err := h.db.QueryRow(`
		SELECT total_clues, start_time FROM game_settings
	`).Scan(&totalClues, &startTime)
	if err != nil {
		totalClues = 1
	}

	// Get groups ordered by completion status and progress
	rows, err := h.db.Query(`
		SELECT id, name, pathway, current_clue_idx, completed, end_time 
		FROM groups 
		ORDER BY completed DESC, current_clue_idx DESC, end_time ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	defer rows.Close()

	var groups []gin.H
	rank := 1
	for rows.Next() {
		var group database.Group
		var endTime sql.NullTime

		err := rows.Scan(
			&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
			&group.Completed, &endTime,
		)
		if err != nil {
			continue
		}

		if endTime.Valid {
			group.EndTime = &endTime.Time
		}

		// Calculate total time if completed
		var totalTime string
		if group.Completed && group.EndTime != nil {
			duration := group.EndTime.Sub(startTime)
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
		if rank == 1 {
			badge = "🥇"
		} else if rank == 2 {
			badge = "🥈"
		} else if rank == 3 {
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
