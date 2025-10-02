package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) QRPage(c *gin.Context) {
	c.HTML(http.StatusOK, "qr.html", nil)
}

func (h *Handler) GetQRData(c *gin.Context) {
	ctx := c.Request.Context()

	// Get filter parameter (optional)
	pathway := c.Query("pathway")

	// Get all clues
	allClues, err := h.clueService.GetAllClues(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clues"})
		return
	}

	// Filter by pathway if specified
	var filteredClues []map[string]interface{}
	for _, clue := range allClues {
		if pathway != "" && clue.Pathway != pathway {
			continue
		}

		filteredClues = append(filteredClues, map[string]interface{}{
			"id":      clue.ID,
			"pathway": clue.Pathway,
			"index":   clue.Index,
			"content": clue.Content,
			"qrcode":  clue.QRCode,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"clues": filteredClues,
	})
}
