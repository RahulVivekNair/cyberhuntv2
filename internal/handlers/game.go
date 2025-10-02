package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
			clueContent = clue.Content
		}
	} else {
		clueContent = "Congratulations! You finished! Check out the leaderboard to see your timing!"
	}

	c.HTML(http.StatusOK, "game.html", gin.H{
		"Group":      group,
		"TotalClues": totalClues,
		"Clue":       clueContent,
	})
}

func (h *Handler) ScanQR(c *gin.Context) {
	groupIDRaw, ok := c.Get("groupID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Group ID missing"})
		return
	}
	groupID, ok := groupIDRaw.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid group ID type"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	totalClues, err := h.gameService.GetTotalClues(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game settings"})
		return
	}

	g, err := h.groupService.ScanAndUpdateProgress(c.Request.Context(), groupID, strings.TrimSpace(req.Code), totalClues)
	if err != nil {
		if strings.Contains(err.Error(), "invalid QR code") {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Wrong QR code"})
			return
		}
		if strings.Contains(err.Error(), "already completed") {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Group already completed"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	if err := h.BroadcastLeaderboard(c.Request.Context()); err != nil {
		// Log the error but don't fail the request since group was added successfully
		c.Header("X-Warning", "Leaderboard broadcast failed")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Correct QR code",
		"group": gin.H{
			"Name":           g.Name,
			"Pathway":        g.Pathway,
			"CurrentClueIdx": g.CurrentClueIdx,
			"Completed":      g.Completed,
		},
	})
}

func (h *Handler) GamePartial(c *gin.Context) {
	groupID, _ := c.Get("groupID")

	group, err := h.groupService.GetGroupByID(c.Request.Context(), groupID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get group"})
		return
	}

	totalClues, err := h.gameService.GetTotalClues(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game settings"})
		return
	}

	var clueContent string
	if !group.Completed {
		clue, err := h.clueService.GetClueByPathwayAndIndex(c.Request.Context(), group.Pathway, group.CurrentClueIdx)
		if err != nil {
			clueContent = "No clue found!"
		} else {
			clueContent = clue.Content
		}
	} else {
		clueContent = "Congratulations! You finished! Check out the leaderboard to see your timing!"
	}

	// Return JSON instead of HTML
	c.JSON(http.StatusOK, gin.H{
		"progress":    fmt.Sprintf("%d/%d", group.CurrentClueIdx, totalClues),
		"completed":   group.Completed,
		"clue":        clueContent,
		"totalClues":  totalClues,
		"currentClue": group.CurrentClueIdx,
	})
}
