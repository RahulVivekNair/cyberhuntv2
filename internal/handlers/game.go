package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GamePage(c *gin.Context) {
	groupID, _ := c.Get("groupID")

	group, err := h.groupService.GetGroupByID(c.Request.Context(), groupID.(int))
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get total clues
	totalClues, _ := h.gameService.GetTotalClues(c.Request.Context())

	// Get current clue if not completed
	var clueContent string
	if !group.Completed {
		clue, err := h.clueService.GetClueByPathwayAndIndex(c.Request.Context(), group.Pathway, group.CurrentClueIdx)
		if err != nil {
			clueContent = "No clue found!"
		} else {
			clueContent = "Clue: " + clue.Content
		}
	} else {
		clueContent = "Congratulations! You finished! Check out the leaderboard to see your timing"
	}

	c.HTML(http.StatusOK, "game.html", gin.H{
		"Group":      group,
		"TotalClues": totalClues,
		"Clue":       clueContent,
	})
}

func (h *Handler) ScanQR(c *gin.Context) {
	groupID, _ := c.Get("groupID")

	group, err := h.groupService.GetGroupByID(c.Request.Context(), groupID.(int))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Group not found"})
		return
	}

	if group.Completed {
		c.JSON(http.StatusOK, gin.H{"message": "Group already completed"})
		return
	}

	// Get scanned QR code from request
	var request struct {
		Code string `json:"code"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get expected QR code for current clue
	expectedClue, err := h.clueService.GetClueByPathwayAndIndex(c.Request.Context(), group.Pathway, group.CurrentClueIdx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Clue not found"})
		return
	}

	// Validate QR code
	if request.Code != expectedClue.QRCode {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Wrong QR code!"})
		return
	}

	// Get total clues
	totalClues, _ := h.gameService.GetTotalClues(c.Request.Context())

	// Update group progress
	newClueIdx := group.CurrentClueIdx + 1
	completed := newClueIdx >= totalClues

	var endTime *time.Time
	if completed {
		t := time.Now().UTC()
		endTime = &t
	}

	err = h.groupService.UpdateGroupProgress(c.Request.Context(), groupID.(int), newClueIdx, completed, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
		return
	}

	// Broadcast updated leaderboard asynchronously so response isn't delayed.
	go func() {
		if err := h.BroadcastLeaderboard(context.Background()); err != nil {
			log.Printf("broadcast leaderboard error: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Correct QR code!"})
}
