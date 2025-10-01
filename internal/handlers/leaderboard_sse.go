package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

//
// ==== Payload Types ====
//

type LeaderboardEntry struct {
	ID             int     `json:"id"`
	Rank           int     `json:"rank"`
	Name           string  `json:"name"`
	Pathway        string  `json:"pathway"`
	CurrentClueIdx int     `json:"current_clue_idx"`
	Completed      bool    `json:"completed"`
	TotalTime      *string `json:"total_time,omitempty"` // null if ongoing
	Badge          string  `json:"badge,omitempty"`
}

type LeaderboardPayload struct {
	Groups      []LeaderboardEntry `json:"groups"`
	TotalClues  int                `json:"totalClues"`
	TotalGroups int                `json:"totalGroups"`
	Completed   int                `json:"completed"`
	InProgress  int                `json:"inProgress"`
}

//
// ==== Hub Implementation ====
//

type LeaderboardHub struct {
	mu        sync.RWMutex
	clients   map[chan []byte]struct{}
	broadcast chan []byte
	cache     []byte
}

func NewLeaderboardHub() *LeaderboardHub {
	h := &LeaderboardHub{
		clients:   make(map[chan []byte]struct{}),
		broadcast: make(chan []byte, 1),
	}
	go h.run()
	return h
}

func (h *LeaderboardHub) run() {
	for payload := range h.broadcast {
		h.mu.Lock()
		h.cache = payload
		for ch := range h.clients {
			select {
			case ch <- payload:
			default: // drop if client is slow
			}
		}
		h.mu.Unlock()
	}
}

func (h *LeaderboardHub) Broadcast(payload []byte) {
	select {
	case h.broadcast <- payload:
	default: // if buffer full, overwrite cache directly
		h.mu.Lock()
		h.cache = payload
		h.mu.Unlock()
	}
}

func (h *LeaderboardHub) AddClient(ctx context.Context) (<-chan []byte, func()) {
	clientCh := make(chan []byte, 1)

	h.mu.Lock()
	h.clients[clientCh] = struct{}{}
	if len(h.cache) > 0 {
		clientCh <- h.cache // send last snapshot immediately
	}
	h.mu.Unlock()

	var once sync.Once
	cancelFn := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.clients, clientCh)
			close(clientCh)
			h.mu.Unlock()
		})
	}

	// Auto-remove client when ctx closes
	go func() {
		<-ctx.Done()
		cancelFn()
	}()

	return clientCh, cancelFn
}

//
// ==== SSE Endpoint ====
//

func (h *Handler) LeaderboardStream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ctx := c.Request.Context()
	clientCh, cancel := h.LeaderboardHub.AddClient(ctx)
	defer cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)

	// Keepalive pings every 15s
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": keepalive\n\n")
			flusher.Flush()
		case payload, ok := <-clientCh:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

//
// ==== Broadcast Builder ====
//

func (h *Handler) BroadcastLeaderboard(ctx context.Context) error {
	totalClues, groups, err := h.groupService.GetLeaderboardData(ctx)
	if err != nil {
		return fmt.Errorf("get leaderboard data: %w", err)
	}

	settings, err := h.gameService.GetGameStatus(ctx)
	var startTime *time.Time
	if err == nil && settings.StartTime != nil {
		startTime = settings.StartTime
	}

	var out []LeaderboardEntry
	rank := 1
	for _, group := range groups {
		entry := LeaderboardEntry{
			ID:             group.ID,
			Rank:           rank,
			Name:           group.Name,
			Pathway:        group.Pathway,
			CurrentClueIdx: group.CurrentClueIdx,
			Completed:      group.Completed,
		}

		// Add total_time if completed
		if group.Completed && group.EndTime != nil && startTime != nil {
			duration := group.EndTime.Sub(*startTime)
			totalSeconds := int(duration.Seconds())
			hh := totalSeconds / 3600
			mm := (totalSeconds % 3600) / 60
			ss := totalSeconds % 60
			formatted := fmt.Sprintf("%02d:%02d:%02d", hh, mm, ss)
			entry.TotalTime = &formatted
		}

		// Assign medal badges for top 3
		switch rank {
		case 1:
			entry.Badge = "🥇"
		case 2:
			entry.Badge = "🥈"
		case 3:
			entry.Badge = "🥉"
		}

		out = append(out, entry)
		rank++
	}

	completed := 0
	for _, group := range groups {
		if group.Completed {
			completed++
		}
	}

	payload := LeaderboardPayload{
		Groups:      out,
		TotalClues:  totalClues,
		TotalGroups: len(groups),
		Completed:   completed,
		InProgress:  len(groups) - completed,
	}

	jsonBytes, _ := json.Marshal(payload)
	h.LeaderboardHub.Broadcast(jsonBytes)
	return nil
}
