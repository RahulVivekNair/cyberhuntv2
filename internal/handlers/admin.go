package handlers

import (
	"cyberhunt/internal/services"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", nil)
}

func (h *Handler) StartGame(c *gin.Context) {
	err := h.gameService.StartGame(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game started successfully!"})
}

func (h *Handler) EndGame(c *gin.Context) {
	err := h.gameService.EndGame(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game ended successfully!"})
}

func (h *Handler) ClearState(c *gin.Context) {
	// Reset game settings
	err := h.gameService.ClearState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear game state"})
		return
	}

	// Reset all groups
	err = h.groupService.ResetGroups(c.Request.Context())
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

	name := strings.TrimSpace(request.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name is required"})
		return
	}

	pathway := strings.ToLower(strings.TrimSpace(request.Pathway))
	if pathway == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pathway is required"})
		return
	}

	validPathways := []string{"red", "blue", "green", "yellow"}
	if !slices.Contains(validPathways, pathway) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid pathway. Must be one of: %s", strings.Join(validPathways, ", ")),
		})
		return
	}

	password := strings.TrimSpace(request.Password)
	if password == "" {
		password = generateRandomPassword(6)
	}

	if err := h.groupService.AddGroup(c.Request.Context(), name, pathway, password); err != nil {
		if errors.Is(err, services.ErrGroupExists) { // <-- check using errors.Is
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add group"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Group added successfully!",
		"password": password,
	})
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	err = h.groupService.DeleteGroup(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully!"})
}

func (h *Handler) GetGameStatus(c *gin.Context) {
	settings, err := h.gameService.GetGameStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get game status"})
		return
	}

	response := gin.H{
		"game_started": settings.GameStarted,
		"game_ended":   settings.GameEnded,
	}
	if settings.StartTime != nil {
		response["start_time"] = settings.StartTime.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetStats(c *gin.Context) {
	totalGroups, completedGroups, inProgressGroups, err := h.groupService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalGroups":      totalGroups,
		"completedGroups":  completedGroups,
		"inProgressGroups": inProgressGroups,
	})
}

func (h *Handler) AdminLeaderboard(c *gin.Context) {
	// Get game settings
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
		// Calculate total time if completed
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

		// Add rank badge (matching user-facing leaderboard)
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
