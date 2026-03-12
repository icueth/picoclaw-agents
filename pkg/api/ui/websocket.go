package ui

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"picoclaw/agent/pkg/logger"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for now - can be restricted in production
			return true
		},
	}
)

// Client represents a single WebSocket connection
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	id         string
	subscribed map[EventType]bool
	filter     Filter
	mu         sync.RWMutex
}

// newClient creates a new WebSocket client
func newClient(hub *Hub, conn *websocket.Conn, id string) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		id:         id,
		subscribed: make(map[EventType]bool),
	}
}

// Subscribe adds an event type to the client's subscription list
func (c *Client) Subscribe(eventType EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribed[eventType] = true
}

// Unsubscribe removes an event type from the client's subscription list
func (c *Client) Unsubscribe(eventType EventType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribed, eventType)
}

// IsSubscribed checks if client is subscribed to an event type
func (c *Client) IsSubscribed(eventType EventType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscribed[eventType] || c.subscribed["*"] // "*" means all events
}

// SetFilter sets the event filter for this client
func (c *Client) SetFilter(filter Filter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filter = filter
}

// ShouldReceiveEvent checks if an event should be sent to this client based on filters
func (c *Client) ShouldReceiveEvent(event *Event) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check event type subscription
	if !c.IsSubscribed(event.Type) {
		return false
	}

	// Apply filters based on event type
	switch payload := event.Payload.(type) {
	case AgentMovedPayload:
		if len(c.filter.AgentIDs) > 0 && !contains(c.filter.AgentIDs, payload.AgentID) {
			return false
		}
		if len(c.filter.RoomIDs) > 0 && !contains(c.filter.RoomIDs, payload.RoomID) {
			return false
		}
	case AgentStatusChangedPayload:
		if len(c.filter.AgentIDs) > 0 && !contains(c.filter.AgentIDs, payload.AgentID) {
			return false
		}
	case TaskCreatedPayload:
		if len(c.filter.AgentIDs) > 0 && !contains(c.filter.AgentIDs, payload.Task.AssigneeID) {
			return false
		}
	case TaskUpdatedPayload:
		if len(c.filter.AgentIDs) > 0 && !contains(c.filter.AgentIDs, payload.Task.AssigneeID) {
			return false
		}
	case RoomOccupantsChangedPayload:
		if len(c.filter.RoomIDs) > 0 && !contains(c.filter.RoomIDs, payload.RoomID) {
			return false
		}
	}

	return true
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (subscriptions, etc.)
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(data []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.DebugF("ui_websocket", map[string]any{
			"error": err.Error(),
		})
		return
	}

	switch msg.Type {
	case "subscribe":
		var req SubscriptionRequest
		payloadBytes, _ := json.Marshal(msg.Payload)
		if err := json.Unmarshal(payloadBytes, &req); err != nil {
			c.sendError("invalid_subscription", "Failed to parse subscription request")
			return
		}
		c.handleSubscribe(req)

	case "unsubscribe":
		var req SubscriptionRequest
		payloadBytes, _ := json.Marshal(msg.Payload)
		if err := json.Unmarshal(payloadBytes, &req); err != nil {
			c.sendError("invalid_unsubscribe", "Failed to parse unsubscribe request")
			return
		}
		c.handleUnsubscribe(req)

	case "ping":
		c.sendPong()

	default:
		// Check if it's a chat event
		if isChatEvent(msg.Type) {
			c.handleChatEvent(msg.Type, data)
		} else {
			logger.DebugF("ui_websocket", map[string]any{
				"message": "Unknown message type",
				"type":    msg.Type,
			})
		}
	}
}

// handleSubscribe processes subscription requests
func (c *Client) handleSubscribe(req SubscriptionRequest) {
	for _, eventType := range req.Events {
		c.Subscribe(eventType)
	}
	c.SetFilter(req.Filter)

	logger.DebugF("ui_websocket", map[string]any{
		"client_id": c.id,
		"events":    req.Events,
		"action":    "subscribed",
	})

	// Send confirmation
	response := map[string]interface{}{
		"type":    "subscribed",
		"events":  req.Events,
		"success": true,
	}
	data, _ := json.Marshal(response)
	c.send <- data
}

// handleUnsubscribe processes unsubscription requests
func (c *Client) handleUnsubscribe(req SubscriptionRequest) {
	for _, eventType := range req.Events {
		c.Unsubscribe(eventType)
	}

	logger.DebugF("ui_websocket", map[string]any{
		"client_id": c.id,
		"events":    req.Events,
		"action":    "unsubscribed",
	})

	// Send confirmation
	response := map[string]interface{}{
		"type":    "unsubscribed",
		"events":  req.Events,
		"success": true,
	}
	data, _ := json.Marshal(response)
	c.send <- data
}

// sendError sends an error message to the client
func (c *Client) sendError(code, message string) {
	response := map[string]interface{}{
		"type": "error",
		"payload": ErrorPayload{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(response)
	select {
	case c.send <- data:
	default:
		// Channel is full, drop the message
	}
}

// sendPong sends a pong response
func (c *Client) sendPong() {
	response := map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now().UTC(),
	}
	data, _ := json.Marshal(response)
	select {
	case c.send <- data:
	default:
		// Channel is full, drop the message
	}
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Event
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			clientCount := len(h.clients)
			h.mu.Unlock()
			logger.DebugF("ui_websocket", map[string]any{
				"client_id":    client.id,
				"total":        clientCount,
				"action":       "connected",
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			clientCount := len(h.clients)
			h.mu.Unlock()
			logger.DebugF("ui_websocket", map[string]any{
				"client_id":    client.id,
				"total":        clientCount,
				"action":       "disconnected",
			})

		case event := <-h.broadcast:
			h.broadcastEvent(event)
		}
	}
}

// broadcastEvent sends an event to all subscribed clients
func (h *Hub) broadcastEvent(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		logger.DebugF("ui_websocket", map[string]any{
			"error": err.Error(),
		})
		return
	}

	for client := range h.clients {
		if client.ShouldReceiveEvent(event) {
			select {
			case client.send <- data:
			default:
				// Client's send channel is full, close and remove
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(event *Event) {
	select {
	case h.broadcast <- event:
	default:
		// Broadcast channel is full, log warning
		logger.DebugF("ui_websocket", map[string]any{
			"warning": "Broadcast channel full, event dropped",
		})
	}
}

// BroadcastAgentMoved broadcasts an agent moved event
func (h *Hub) BroadcastAgentMoved(agentID, roomID string, from, to Position) {
	h.Broadcast(&Event{
		Type:      EventAgentMoved,
		Timestamp: time.Now(),
		Payload: AgentMovedPayload{
			AgentID:   agentID,
			RoomID:    roomID,
			From:      from,
			To:        to,
			Timestamp: time.Now(),
		},
	})
}

// BroadcastAgentStatusChanged broadcasts an agent status change event
func (h *Hub) BroadcastAgentStatusChanged(agentID string, oldStatus, newStatus AgentStatus, reason string) {
	h.Broadcast(&Event{
		Type:      EventAgentStatusChanged,
		Timestamp: time.Now(),
		Payload: AgentStatusChangedPayload{
			AgentID:   agentID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Reason:    reason,
		},
	})
}

// BroadcastTaskCreated broadcasts a task created event
func (h *Hub) BroadcastTaskCreated(task Task, creatorID string) {
	h.Broadcast(&Event{
		Type:      EventTaskCreated,
		Timestamp: time.Now(),
		Payload: TaskCreatedPayload{
			Task:      task,
			CreatorID: creatorID,
		},
	})
}

// BroadcastTaskUpdated broadcasts a task updated event
func (h *Hub) BroadcastTaskUpdated(task Task, changes []string, updatedBy string) {
	h.Broadcast(&Event{
		Type:      EventTaskUpdated,
		Timestamp: time.Now(),
		Payload: TaskUpdatedPayload{
			Task:      task,
			Changes:   changes,
			UpdatedBy: updatedBy,
		},
	})
}

// BroadcastTaskAssigned broadcasts a task assigned event
func (h *Hub) BroadcastTaskAssigned(taskID, assigneeID, assignerID string) {
	h.Broadcast(&Event{
		Type:      EventTaskAssigned,
		Timestamp: time.Now(),
		Payload: TaskAssignedPayload{
			TaskID:     taskID,
			AssigneeID: assigneeID,
			AssignerID: assignerID,
		},
	})
}

// BroadcastRoomOccupantsChanged broadcasts a room occupants change event
func (h *Hub) BroadcastRoomOccupantsChanged(roomID string, occupants, joined, left []string) {
	h.Broadcast(&Event{
		Type:      EventRoomOccupantsChanged,
		Timestamp: time.Now(),
		Payload: RoomOccupantsChangedPayload{
			RoomID:    roomID,
			Occupants: occupants,
			Joined:    joined,
			Left:      left,
		},
	})
}

// BroadcastSystemStatus broadcasts a system status event
func (h *Hub) BroadcastSystemStatus(onlineAgents, activeTasks int, memoryUsage uint64, uptime string) {
	h.Broadcast(&Event{
		Type:      EventSystemStatus,
		Timestamp: time.Now(),
		Payload: SystemStatusPayload{
			OnlineAgents: onlineAgents,
			ActiveTasks:  activeTasks,
			MemoryUsage:  memoryUsage,
			Uptime:       uptime,
		},
	})
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// WebSocketHandler handles WebSocket upgrade requests
func (h *Hub) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.DebugF("ui_websocket", map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Generate client ID from request or create new one
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = generateClientID()
	}

	client := newClient(h, conn, clientID)

	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isChatEvent checks if a message type is a chat event
func isChatEvent(msgType string) bool {
	chatEvents := []string{
		"chat_message",
		"typing_start",
		"typing_stop",
		"session_joined",
		"session_left",
		"agent_status",
		"settings_saved",
	}
	for _, t := range chatEvents {
		if t == msgType {
			return true
		}
	}
	return false
}
