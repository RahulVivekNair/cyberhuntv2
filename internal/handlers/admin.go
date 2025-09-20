package handlers

import (
	"cyberhunt/internal/services"
	"cyberhunt/internal/utils"
	"database/sql"
	"errors"
	"fmt"
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
		if errors.Is(err, services.ErrGameAlreadyStarted) {
			c.JSON(http.StatusConflict, gin.H{"error": "Game already started"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game started successfully!"})
}

func (h *Handler) EndGame(c *gin.Context) {
	err := h.gameService.EndGame(c.Request.Context())
	if err != nil {
		switch err {
		case services.ErrGameNotStarted:
			// Game hasn't started yet
			c.JSON(http.StatusBadRequest, gin.H{"error": "Game has not started yet"})
		case services.ErrGameAlreadyEnded:
			// Game already ended
			c.JSON(http.StatusConflict, gin.H{"error": "Game already ended"})
		default:
			// Other unexpected errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end game"})
		}
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{"message": "Game ended successfully!"})
}

func (h *Handler) ClearState(c *gin.Context) {
	err := h.gameService.ClearAllState(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear state"})
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
		password = utils.GenerateRandomPassword(6)
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
	if err != nil || groupID <= 0 { // Add bounds checking
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	err = h.groupService.DeleteGroup(c.Request.Context(), groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	// Or, if you prefer to keep a success message:
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
