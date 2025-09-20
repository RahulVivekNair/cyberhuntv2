package handlers

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SeedPage(c *gin.Context) {
	c.HTML(http.StatusOK, "seed.html", nil)
}

func (h *Handler) SeedGroups(c *gin.Context) {
	pathways := []string{"red", "blue", "yellow", "green"}
	groupsPerPathway := 12 // 50 / 4 = 12.5, so 12 or 13. Let's do 13 for one pathway.

	pathwayCount := make(map[string]int)
	for i := 0; i < 4; i++ {
		pathwayCount[pathways[i]] = groupsPerPathway
	}

	for pathway, count := range pathwayCount {
		for i := 0; i < count; i++ {
			name := fmt.Sprintf("Group_%s_%03d", pathway, i+1)
			password := "test"

			err := h.groupService.AddGroup(c.Request.Context(), name, pathway, password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed groups"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Groups seeded successfully!"})
}

func (h *Handler) SeedClues(c *gin.Context) {
	pathways := []string{"red", "blue", "yellow", "green"}
	cluesPerPathway := 10

	riddles := []string{
		"What has keys but can't open locks?",
		"What gets wetter as it dries?",
		"What has a head, a tail, but no body?",
		"What has one eye but can't see?",
		"What can travel around the world while staying in a corner?",
		"What has hands but can't clap?",
		"What has a neck but no head?",
		"What has a bottom at the top?",
		"What has many teeth but can't bite?",
		"What has a ring but no finger?",
		"What has a spine but no bones?",
		"What has branches but no leaves?",
		"What has a face and two hands but no arms or legs?",
		"What has a foot but no legs?",
		"What has a bed but doesn't sleep?",
		"What has a tongue but can't taste?",
		"What has a heart that doesn't beat?",
		"What has a thumb and four fingers but isn't alive?",
		"What has a mouth but doesn't eat?",
		"What has a nose but doesn't smell?",
		"What has a back but no front?",
		"What has a head but no brain?",
		"What has a crown but isn't a king?",
		"What has a stem but no petals?",
		"What has a cap but no head?",
		"What has a belt but no pants?",
		"What has a bow but no arrow?",
		"What has a bridge but no water?",
		"What has a button but no shirt?",
		"What has a chain but no lock?",
		"What has a chimney but no fire?",
		"What has a clock but no hands?",
		"What has a comb but no hair?",
		"What has a crown but no jewels?",
		"What has a door but no house?",
		"What has a drum but no sticks?",
		"What has a eye but no vision?",
		"What has a face but no features?",
		"What has a foot but no shoe?",
		"What has a handle but no door?",
	}

	// Clear existing clues
	err := h.clueService.ClearClues(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing clues"})
		return
	}

	for _, pathway := range pathways {
		for i := 0; i < cluesPerPathway; i++ {
			qrCode := fmt.Sprintf("%s_%03d", pathway, i)
			content := riddles[rand.Intn(len(riddles))]

			err := h.clueService.AddClue(c.Request.Context(), pathway, i, content, qrCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to insert clue %s_%03d: %v", pathway, i, err)})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clues seeded successfully!"})
}

func (h *Handler) UpdateTotalClues(c *gin.Context) {
	var request struct {
		TotalClues int `json:"total_clues"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if request.TotalClues <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Total clues must be greater than 0"})
		return
	}

	err := h.adminService.UpdateTotalClues(c.Request.Context(), request.TotalClues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update total clues"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Total clues updated successfully!"})
}
