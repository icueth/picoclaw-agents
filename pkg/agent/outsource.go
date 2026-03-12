package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"picoclaw/agent/pkg/logger"
)

// OutsourceAgent represents a temporary outsourced agent in the pool
type OutsourceAgent struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	ParentID  string    `json:"parent_id"`
	TaskID    string    `json:"task_id"`
	HiredAt   time.Time `json:"hired_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Model     string    `json:"model"`
}

// IsExpired checks if the outsource agent has expired
func (oa *OutsourceAgent) IsExpired() bool {
	return time.Now().After(oa.ExpiresAt)
}

// TimeUntilExpiry returns the duration until the agent expires
func (oa *OutsourceAgent) TimeUntilExpiry() time.Duration {
	return time.Until(oa.ExpiresAt)
}

// OutsourcePool manages temporary outsourced agents with capacity limits and expiration
type OutsourcePool struct {
	activeAgents map[string]*OutsourceAgent
	maxPoolSize  int
	defaultTTL   time.Duration
	mu           sync.RWMutex
}

// NewOutsourcePool creates a new outsource pool with specified capacity
func NewOutsourcePool(maxSize int, defaultTTL time.Duration) *OutsourcePool {
	if maxSize <= 0 {
		maxSize = 10 // Default pool size
	}
	if defaultTTL <= 0 {
		defaultTTL = 30 * time.Minute // Default 30 minute TTL
	}

	pool := &OutsourcePool{
		activeAgents: make(map[string]*OutsourceAgent),
		maxPoolSize:  maxSize,
		defaultTTL:   defaultTTL,
	}

	// Start background cleanup goroutine
	go pool.cleanupLoop()

	return pool
}

// Hire creates a new temporary outsource agent and adds it to the pool
// Returns the agent ID and nil error on success, or empty string and error on failure
func (p *OutsourcePool) Hire(parentID, role, taskID string) (*OutsourceAgent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if pool is at capacity
	if len(p.activeAgents) >= p.maxPoolSize {
		logger.WarnCF("outsource", "Pool is at capacity, cannot hire new agent", map[string]any{
			"pool_size":     len(p.activeAgents),
			"max_pool_size": p.maxPoolSize,
			"parent_id":     parentID,
			"role":          role,
		})
		return nil, fmt.Errorf("outsource pool is at capacity (%d/%d)", len(p.activeAgents), p.maxPoolSize)
	}

	// Clean up expired agents before adding new one
	p.cleanupExpiredLocked()

	// Generate unique agent ID
	agentID := fmt.Sprintf("outsource-%s", uuid.New().String()[:8])

	now := time.Now()
	agent := &OutsourceAgent{
		ID:        agentID,
		Role:      role,
		ParentID:  parentID,
		TaskID:    taskID,
		HiredAt:   now,
		ExpiresAt: now.Add(p.defaultTTL),
		Model:     "", // Can be set later if needed
	}

	p.activeAgents[agentID] = agent

	logger.InfoCF("outsource", "Hired new outsource agent", map[string]any{
		"agent_id":   agentID,
		"parent_id":  parentID,
		"role":       role,
		"task_id":    taskID,
		"expires_at": agent.ExpiresAt,
		"pool_size":  len(p.activeAgents),
	})

	return agent, nil
}

// HireWithOptions creates a new outsource agent with custom options
func (p *OutsourcePool) HireWithOptions(parentID, role, taskID, model string, ttl time.Duration) (*OutsourceAgent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.activeAgents) >= p.maxPoolSize {
		return nil, fmt.Errorf("outsource pool is at capacity (%d/%d)", len(p.activeAgents), p.maxPoolSize)
	}

	p.cleanupExpiredLocked()

	agentID := fmt.Sprintf("outsource-%s", uuid.New().String()[:8])

	if ttl <= 0 {
		ttl = p.defaultTTL
	}

	now := time.Now()
	agent := &OutsourceAgent{
		ID:        agentID,
		Role:      role,
		ParentID:  parentID,
		TaskID:    taskID,
		HiredAt:   now,
		ExpiresAt: now.Add(ttl),
		Model:     model,
	}

	p.activeAgents[agentID] = agent

	logger.InfoCF("outsource", "Hired outsource agent with custom options", map[string]any{
		"agent_id":   agentID,
		"parent_id":  parentID,
		"role":       role,
		"task_id":    taskID,
		"model":      model,
		"ttl":        ttl.String(),
		"expires_at": agent.ExpiresAt,
		"pool_size":  len(p.activeAgents),
	})

	return agent, nil
}

// Release removes an outsource agent from the pool by ID
// Returns true if the agent was found and removed, false otherwise
func (p *OutsourcePool) Release(agentID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, exists := p.activeAgents[agentID]
	if !exists {
		logger.WarnCF("outsource", "Attempted to release non-existent agent", map[string]any{
			"agent_id": agentID,
		})
		return false
	}

	delete(p.activeAgents, agentID)

	logger.InfoCF("outsource", "Released outsource agent", map[string]any{
		"agent_id":  agentID,
		"role":      agent.Role,
		"parent_id": agent.ParentID,
		"hired_at":  agent.HiredAt,
		"pool_size": len(p.activeAgents),
	})

	return true
}

// GetActiveAgents returns a list of all active (non-expired) outsource agents
func (p *OutsourcePool) GetActiveAgents() []*OutsourceAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Clean up expired agents first
	p.cleanupExpiredLocked()

	agents := make([]*OutsourceAgent, 0, len(p.activeAgents))
	for _, agent := range p.activeAgents {
		agents = append(agents, agent)
	}

	return agents
}

// GetAgent returns a specific agent by ID
func (p *OutsourcePool) GetAgent(agentID string) (*OutsourceAgent, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, exists := p.activeAgents[agentID]
	if !exists {
		return nil, false
	}

	// Check if expired
	if agent.IsExpired() {
		return nil, false
	}

	return agent, true
}

// IsFull returns true if the pool has reached its maximum capacity
func (p *OutsourcePool) IsFull() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.activeAgents) >= p.maxPoolSize
}

// GetPoolSize returns the current number of active agents in the pool
func (p *OutsourcePool) GetPoolSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.activeAgents)
}

// GetMaxPoolSize returns the maximum capacity of the pool
func (p *OutsourcePool) GetMaxPoolSize() int {
	return p.maxPoolSize
}

// GetAvailableSlots returns the number of available slots in the pool
func (p *OutsourcePool) GetAvailableSlots() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	available := p.maxPoolSize - len(p.activeAgents)
	if available < 0 {
		return 0
	}
	return available
}

// ExtendExpiration extends the expiration time of an agent
func (p *OutsourcePool) ExtendExpiration(agentID string, additionalDuration time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, exists := p.activeAgents[agentID]
	if !exists {
		return false
	}

	agent.ExpiresAt = agent.ExpiresAt.Add(additionalDuration)

	logger.InfoCF("outsource", "Extended agent expiration", map[string]any{
		"agent_id":     agentID,
		"new_expires":  agent.ExpiresAt,
		"extension":    additionalDuration.String(),
	})

	return true
}

// GetAgentsByParent returns all agents hired by a specific parent
func (p *OutsourcePool) GetAgentsByParent(parentID string) []*OutsourceAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var agents []*OutsourceAgent
	for _, agent := range p.activeAgents {
		if agent.ParentID == parentID && !agent.IsExpired() {
			agents = append(agents, agent)
		}
	}

	return agents
}

// GetAgentsByRole returns all agents with a specific role
func (p *OutsourcePool) GetAgentsByRole(role string) []*OutsourceAgent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var agents []*OutsourceAgent
	for _, agent := range p.activeAgents {
		if agent.Role == role && !agent.IsExpired() {
			agents = append(agents, agent)
		}
	}

	return agents
}

// cleanupExpiredLocked removes expired agents from the pool
// Must be called with lock held
func (p *OutsourcePool) cleanupExpiredLocked() {
	now := time.Now()
	for id, agent := range p.activeAgents {
		if now.After(agent.ExpiresAt) {
			delete(p.activeAgents, id)
			logger.InfoCF("outsource", "Cleaned up expired agent", map[string]any{
				"agent_id":   id,
				"role":       agent.Role,
				"parent_id":  agent.ParentID,
				"expired_at": agent.ExpiresAt,
			})
		}
	}
}

// cleanupLoop runs periodically to clean up expired agents
func (p *OutsourcePool) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		expiredCount := 0
		now := time.Now()
		for id, agent := range p.activeAgents {
			if now.After(agent.ExpiresAt) {
				delete(p.activeAgents, id)
				expiredCount++
			}
		}
		poolSize := len(p.activeAgents)
		p.mu.Unlock()

		if expiredCount > 0 {
			logger.InfoCF("outsource", "Periodic cleanup completed", map[string]any{
				"expired_count": expiredCount,
				"pool_size":     poolSize,
			})
		}
	}
}

// GetPoolStats returns statistics about the pool
func (p *OutsourcePool) GetPoolStats() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var totalAge time.Duration
	var expiredCount int
	now := time.Now()

	for _, agent := range p.activeAgents {
		totalAge += now.Sub(agent.HiredAt)
		if now.After(agent.ExpiresAt) {
			expiredCount++
		}
	}

	avgAge := time.Duration(0)
	if len(p.activeAgents) > 0 {
		avgAge = totalAge / time.Duration(len(p.activeAgents))
	}

	return map[string]any{
		"pool_size":       len(p.activeAgents),
		"max_pool_size":   p.maxPoolSize,
		"available_slots": p.maxPoolSize - len(p.activeAgents),
		"is_full":         len(p.activeAgents) >= p.maxPoolSize,
		"expired_count":   expiredCount,
		"average_age":     avgAge.String(),
		"default_ttl":     p.defaultTTL.String(),
	}
}

// SetMaxPoolSize updates the maximum pool size
// Note: This does not affect existing agents, only future hires
func (p *OutsourcePool) SetMaxPoolSize(newSize int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if newSize > 0 {
		p.maxPoolSize = newSize
		logger.InfoCF("outsource", "Updated max pool size", map[string]any{
			"new_max_size": newSize,
		})
	}
}

// SetDefaultTTL updates the default time-to-live for new agents
func (p *OutsourcePool) SetDefaultTTL(newTTL time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if newTTL > 0 {
		p.defaultTTL = newTTL
		logger.InfoCF("outsource", "Updated default TTL", map[string]any{
			"new_ttl": newTTL.String(),
		})
	}
}

// Clear removes all agents from the pool
func (p *OutsourcePool) Clear() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	count := len(p.activeAgents)
	p.activeAgents = make(map[string]*OutsourceAgent)

	logger.InfoCF("outsource", "Cleared all agents from pool", map[string]any{
		"cleared_count": count,
	})

	return count
}
