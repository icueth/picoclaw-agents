package mailbox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMailbox(t *testing.T) {
	mb := NewMailbox("agent-1", 100)
	
	assert.Equal(t, "agent-1", mb.agentID)
	assert.Equal(t, 100, mb.capacity)
	assert.Equal(t, 0, mb.Size())
	assert.Equal(t, 0, mb.GetUnreadCount())
}

func TestMailbox_SendAndReceive(t *testing.T) {
	mb := NewMailbox("agent-1", 100)
	
	// Send message
	msg := Message{
		ID:        "msg-1",
		Type:      MessageTypeTask,
		From:      "jarvis",
		To:        "agent-1",
		Priority:  PriorityNormal,
		Content:   "Test task",
		CreatedAt: time.Now(),
	}
	
	err := mb.Send(msg)
	require.NoError(t, err)
	assert.Equal(t, 1, mb.Size())
	assert.Equal(t, 1, mb.GetUnreadCount())
	
	// Receive message
	received, err := mb.Receive()
	require.NoError(t, err)
	assert.Equal(t, "msg-1", received.ID)
	assert.Equal(t, "Test task", received.Content)
	assert.True(t, received.Read)
	assert.Equal(t, 0, mb.GetUnreadCount())
}

func TestMailbox_Priority(t *testing.T) {
	mb := NewMailbox("agent-1", 100)
	
	// Send messages with different priorities
	messages := []Message{
		{ID: "msg-3", Priority: PriorityLow, Content: "Low priority", CreatedAt: time.Now()},
		{ID: "msg-1", Priority: PriorityCritical, Content: "Critical", CreatedAt: time.Now()},
		{ID: "msg-2", Priority: PriorityHigh, Content: "High priority", CreatedAt: time.Now()},
		{ID: "msg-4", Priority: PriorityNormal, Content: "Normal", CreatedAt: time.Now()},
	}
	
	for _, msg := range messages {
		err := mb.Send(msg)
		require.NoError(t, err)
	}
	
	// Receive should return in priority order
	received1, _ := mb.Receive()
	assert.Equal(t, PriorityCritical, received1.Priority)
	
	received2, _ := mb.Receive()
	assert.Equal(t, PriorityHigh, received2.Priority)
	
	received3, _ := mb.Receive()
	assert.Equal(t, PriorityNormal, received3.Priority)
	
	received4, _ := mb.Receive()
	assert.Equal(t, PriorityLow, received4.Priority)
}

func TestMailbox_Capacity(t *testing.T) {
	mb := NewMailbox("agent-1", 3)
	
	// Fill mailbox
	for i := 0; i < 3; i++ {
		msg := Message{
			ID:        string(rune('0' + i)),
			Priority:  PriorityNormal,
			Content:   "Message",
			CreatedAt: time.Now(),
		}
		err := mb.Send(msg)
		require.NoError(t, err)
	}
	
	assert.Equal(t, 3, mb.Size())
	
	// Send critical message - should replace lowest priority
	criticalMsg := Message{
		ID:        "critical",
		Priority:  PriorityCritical,
		Content:   "Critical message",
		CreatedAt: time.Now(),
	}
	
	err := mb.Send(criticalMsg)
	require.NoError(t, err)
	assert.Equal(t, 3, mb.Size())
	
	// Verify critical message is there
	received, _ := mb.Receive()
	assert.Equal(t, "critical", received.ID)
}

func TestMailbox_Subscribe(t *testing.T) {
	mb := NewMailbox("agent-1", 100)
	
	// Subscribe
	ch := mb.Subscribe()
	require.NotNil(t, ch)
	
	// Send message
	go func() {
		msg := Message{
			ID:        "msg-1",
			Type:      MessageTypeTask,
			From:      "jarvis",
			Content:   "Test",
			CreatedAt: time.Now(),
		}
		mb.Send(msg)
	}()
	
	// Wait for notification
	select {
	case msg := <-ch:
		assert.Equal(t, "msg-1", msg.ID)
	case <-time.After(time.Second):
		t.Fatal("Did not receive message notification")
	}
}

func TestHub_RegisterAndGet(t *testing.T) {
	hub := NewHub(100)
	
	// Register agent
	mb := hub.Register("agent-1")
	require.NotNil(t, mb)
	assert.Equal(t, "agent-1", mb.agentID)
	
	// Get registered agent
	retrieved, ok := hub.Get("agent-1")
	assert.True(t, ok)
	assert.Equal(t, mb, retrieved)
	
	// Get non-existent agent
	_, ok = hub.Get("non-existent")
	assert.False(t, ok)
}

func TestHub_SendTo(t *testing.T) {
	hub := NewHub(100)
	
	// Register sender and receiver
	hub.Register("jarvis")
	hub.Register("agent-1")
	
	// Send message
	msg := Message{
		ID:        "msg-1",
		Type:      MessageTypeTask,
		From:      "jarvis",
		To:        "agent-1",
		Priority:  PriorityNormal,
		Content:   "Test task",
		CreatedAt: time.Now(),
	}
	
	err := hub.SendTo("agent-1", msg)
	require.NoError(t, err)
	
	// Verify message received
	mb, _ := hub.Get("agent-1")
	assert.Equal(t, 1, mb.GetUnreadCount())
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub(100)
	
	// Register multiple agents
	hub.Register("jarvis")
	hub.Register("agent-1")
	hub.Register("agent-2")
	hub.Register("agent-3")
	
	// Broadcast message
	msg := Message{
		ID:        "broadcast-1",
		Type:      MessageTypeBroadcast,
		From:      "jarvis",
		To:        "all",
		Content:   "Hello everyone",
		CreatedAt: time.Now(),
	}
	
	failed := hub.Broadcast(msg)
	assert.Empty(t, failed)
	
	// Verify all agents received message
	for _, id := range []string{"agent-1", "agent-2", "agent-3"} {
		mb, ok := hub.Get(id)
		require.True(t, ok)
		assert.Equal(t, 1, mb.GetUnreadCount(), "Agent %s should have message", id)
	}
	
	// Sender should not receive their own broadcast
	jarvisMailbox, _ := hub.Get("jarvis")
	assert.Equal(t, 0, jarvisMailbox.GetUnreadCount())
}

func TestHub_Cleanup(t *testing.T) {
	hub := NewHub(100)
	
	hub.Register("agent-1")
	
	// Send message with short expiry
	expiry := time.Now().Add(100 * time.Millisecond)
	msg := Message{
		ID:        "expiring",
		Type:      MessageTypeTask,
		From:      "jarvis",
		Priority:  PriorityNormal,
		Content:   "This will expire",
		CreatedAt: time.Now(),
		ExpiresAt: &expiry,
	}
	
	mb, _ := hub.Get("agent-1")
	mb.Send(msg)
	assert.Equal(t, 1, mb.Size())
	
	// Wait for expiry
	time.Sleep(200 * time.Millisecond)
	
	// Cleanup
	removed := mb.CleanupExpired()
	assert.Equal(t, 1, removed)
	assert.Equal(t, 0, mb.Size())
}

func TestHub_StartCleanup(t *testing.T) {
	hub := NewHub(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	hub.Register("agent-1")
	
	// Start cleanup goroutine
	hub.StartCleanup(ctx, 100*time.Millisecond)
	
	// Send expiring message
	expiry := time.Now().Add(150 * time.Millisecond)
	msg := Message{
		ID:        "expiring",
		Type:      MessageTypeTask,
		From:      "jarvis",
		Priority:  PriorityNormal,
		Content:   "Auto cleanup test",
		CreatedAt: time.Now(),
		ExpiresAt: &expiry,
	}
	
	mb, _ := hub.Get("agent-1")
	mb.Send(msg)
	
	// Wait for auto cleanup
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, 0, mb.Size())
}

func TestGenerateMessageID(t *testing.T) {
	id1 := GenerateMessageID("agent-1")
	id2 := GenerateMessageID("agent-1")
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "agent-1")
}
