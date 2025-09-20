package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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
			default:
			}
		}
		h.mu.Unlock()
	}
}

func (h *LeaderboardHub) Broadcast(payload []byte) {
	select {
	case h.broadcast <- payload:
	default:
		h.mu.Lock()
		h.cache = payload
		h.mu.Unlock()
	}
}

func (h *LeaderboardHub) AddClient(ctx context.Context) (ch <-chan []byte, cancel func()) {
	clientCh := make(chan []byte, 1)

	h.mu.Lock()
	h.clients[clientCh] = struct{}{}
	if len(h.cache) > 0 {
		clientCh <- h.cache
	}
	h.mu.Unlock()

	done := make(chan struct{})
	closed := false

	cancelFn := func() {
		h.mu.Lock()
		if _, ok := h.clients[clientCh]; ok && !closed {
			delete(h.clients, clientCh)
			close(clientCh)
			closed = true
		}
		h.mu.Unlock()

		if !closed {
			close(done)
			closed = true
		}
	}

	go func() {
		select {
		case <-ctx.Done():
			cancelFn()
		case <-done:
		}
	}()

	return clientCh, cancelFn
}

// SSE endpoint
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

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-clientCh:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

// BroadcastLeaderboard builds leaderboard JSON and sends it
func (h *Handler) BroadcastLeaderboard(ctx context.Context) error {
	totalClues, groups, err := h.groupService.GetLeaderboardData(ctx)
	if err != nil {
		return errors.New("failed to fetch leaderboard") // generic for client
	}

	// Get game start time for calculating total_time
	settings, err := h.gameService.GetGameStatus(ctx)
	var startTime *time.Time
	if err == nil && settings.StartTime != nil {
		startTime = settings.StartTime
	}

	var out []map[string]interface{}
	rank := 1
	for _, group := range groups {
		entry := map[string]interface{}{
			"rank":             rank,
			"name":             group.Name,
			"pathway":          group.Pathway,
			"current_clue_idx": group.CurrentClueIdx,
			"completed":        group.Completed,
		}

		// Calculate total_time for completed groups
		if group.Completed && group.EndTime != nil && startTime != nil {
			duration := group.EndTime.Sub(*startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			entry["total_time"] = fmt.Sprintf("%d:%02d", minutes, seconds)
		} else {
			entry["total_time"] = nil
		}

		switch rank {
		case 1:
			entry["badge"] = "🥇"
		case 2:
			entry["badge"] = "🥈"
		case 3:
			entry["badge"] = "🥉"
		}
		out = append(out, entry)
		rank++
	}

	payload, _ := json.Marshal(gin.H{
		"groups":     out,
		"totalClues": totalClues,
	})

	h.LeaderboardHub.Broadcast(payload)
	return nil
}
