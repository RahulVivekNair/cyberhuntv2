package handlers

import (
	"cyberhunt/internal/database"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", nil)
}

func (h *Handler) StartGame(c *gin.Context) {
	_, err := h.db.Exec(`
		UPDATE game_settings 
		SET game_started = TRUE, start_time = datetime('now')
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game started successfully!"})
}

func (h *Handler) EndGame(c *gin.Context) {
	_, err := h.db.Exec(`
		UPDATE game_settings SET game_ended = TRUE
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game ended successfully!"})
}

func (h *Handler) ClearState(c *gin.Context) {
	// Reset game settings
	_, err := h.db.Exec(`
		UPDATE game_settings 
		SET game_started = FALSE, game_ended = FALSE, start_time = NULL
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear game state"})
		return
	}

	// Reset all groups
	_, err = h.db.Exec(`
		UPDATE groups 
		SET current_clue_idx = 0, completed = FALSE, end_time = NULL
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game state cleared successfully!"})
}

func (h *Handler) AddGroup(c *gin.Context) {
	var request struct {
		Name     string `json:"name"`
		Pathway  string `json:"pathway"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Generate password if not provided
	password := request.Password
	if password == "" {
		password = generateRandomPassword(6)
	}

	_, err := h.db.Exec(`
		INSERT INTO groups (name, pathway, password) 
		VALUES (?, ?, ?)
	`, request.Name, request.Pathway, password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Group added successfully!",
		"password": password,
	})
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")

	_, err := h.db.Exec("DELETE FROM groups WHERE id = ?", groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully!"})
}

func (h *Handler) GetStats(c *gin.Context) {
	var totalGroups, completedGroups, inProgressGroups int

	// Get total groups
	err := h.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&totalGroups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	// Get completed groups
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM groups WHERE completed = TRUE
	`).Scan(&completedGroups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	inProgressGroups = totalGroups - completedGroups

	c.JSON(http.StatusOK, gin.H{
		"totalGroups":      totalGroups,
		"completedGroups":  completedGroups,
		"inProgressGroups": inProgressGroups,
	})
}

func (h *Handler) AdminLeaderboard(c *gin.Context) {
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

		// Add rank badge (matching user-facing leaderboard)
		var badge string
		if rank == 1 {
			badge = "🥇"
		} else if rank == 2 {
			badge = "🥈"
		} else if rank == 3 {
			badge = "🥉"
		}

		groups = append(groups, gin.H{
			"id":               group.ID,
			"rank":             rank,
			"badge":            badge,
			"name":             group.Name,
			"pathway":          group.Pathway,
			"current_clue_idx": group.CurrentClueIdx,
			"completed":        group.Completed,
			"total_time":       totalTime,
			"progress_percent": int(float64(group.CurrentClueIdx) / float64(totalClues) * 100),
		})

		rank++
	}

	c.JSON(http.StatusOK, gin.H{
		"groups":     groups,
		"totalClues": totalClues,
	})
}

func generateRandomPassword(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
