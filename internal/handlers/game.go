package handlers

import (
	"cyberhunt/internal/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GamePage(c *gin.Context) {
	groupID, _ := c.Get("groupID")

	var group database.Group
	err := h.db.QueryRow(`
		SELECT id, name, pathway, current_clue_idx, completed, end_time 
		FROM groups WHERE id = ?
	`, groupID).Scan(
		&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
		&group.Completed, &group.EndTime,
	)

	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get total clues
	var totalClues int
	err = h.db.QueryRow("SELECT total_clues FROM game_settings").Scan(&totalClues)
	if err != nil {
		totalClues = 1
	}

	// Get current clue if not completed
	var clueContent string
	if !group.Completed {
		var clue database.Clue
		err = h.db.QueryRow(`
			SELECT content FROM clues WHERE pathway = ? AND index_num = ?
		`, group.Pathway, group.CurrentClueIdx).Scan(&clue.Content)

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

	var group database.Group
	err := h.db.QueryRow(`
		SELECT id, name, pathway, current_clue_idx, completed 
		FROM groups WHERE id = ?
	`, groupID).Scan(
		&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx, &group.Completed,
	)

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
	var expectedCode string
	err = h.db.QueryRow(`
		SELECT qrcode FROM clues WHERE pathway = ? AND index_num = ?
	`, group.Pathway, group.CurrentClueIdx).Scan(&expectedCode)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Clue not found"})
		return
	}

	// Validate QR code
	if request.Code != expectedCode {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Wrong QR code!"})
		return
	}

	// Get total clues
	var totalClues int
	err = h.db.QueryRow("SELECT total_clues FROM game_settings").Scan(&totalClues)
	if err != nil {
		totalClues = 1
	}

	// Update group progress
	newClueIdx := group.CurrentClueIdx + 1
	completed := newClueIdx >= totalClues

	var endTime interface{}
	if completed {
		t := time.Now().UTC()
		endTime = t
	} else {
		endTime = nil
	}

	_, err = h.db.Exec(`
    UPDATE groups 
    SET current_clue_idx = ?, completed = ?, end_time = ?
    WHERE id = ?
`, newClueIdx, completed, endTime, groupID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Correct QR code!"})
}
