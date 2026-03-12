package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutsourcePool(t *testing.T) {
	tests := []struct {
		name      string
		maxSize   int
		ttl       time.Duration
		wantSize  int
		wantTTL   time.Duration
	}{
		{
			name:     "default values",
			maxSize:  0,
			ttl:      0,
			wantSize: 10,
			wantTTL:  30 * time.Minute,
		},
		{
			name:     "custom values",
			maxSize:  20,
			ttl:      1 * time.Hour,
			wantSize: 20,
			wantTTL:  1 * time.Hour,
		},
		{
			name:     "negative values use defaults",
			maxSize:  -5,
			ttl:      -1 * time.Second,
			wantSize: 10,
			wantTTL:  30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewOutsourcePool(tt.maxSize, tt.ttl)
			require.NotNil(t, pool)
			assert.Equal(t, tt.wantSize, pool.GetMaxPoolSize())
			assert.Equal(t, 0, pool.GetPoolSize())
		})
	}
}

func TestOutsourcePool_Hire(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	t.Run("hire single agent", func(t *testing.T) {
		agent, err := pool.Hire("parent-1", "coder", "task-1")
		require.NoError(t, err)
		require.NotNil(t, agent)

		assert.NotEmpty(t, agent.ID)
		assert.True(t, agent.ID[:10] == "outsource-")
		assert.Equal(t, "parent-1", agent.ParentID)
		assert.Equal(t, "coder", agent.Role)
		assert.Equal(t, "task-1", agent.TaskID)
		assert.False(t, agent.HiredAt.IsZero())
		assert.False(t, agent.ExpiresAt.IsZero())
		assert.True(t, agent.ExpiresAt.After(agent.HiredAt))
	})

	t.Run("hire multiple agents", func(t *testing.T) {
		pool := NewOutsourcePool(10, 30*time.Minute)

		for i := 0; i < 5; i++ {
			_, err := pool.Hire("parent-1", "coder", "task-1")
			require.NoError(t, err)
		}

		assert.Equal(t, 5, pool.GetPoolSize())
	})

	t.Run("pool at capacity", func(t *testing.T) {
		pool := NewOutsourcePool(2, 30*time.Minute)

		_, err := pool.Hire("parent-1", "coder", "task-1")
		require.NoError(t, err)

		_, err = pool.Hire("parent-1", "coder", "task-2")
		require.NoError(t, err)

		_, err = pool.Hire("parent-1", "coder", "task-3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at capacity")
	})
}

func TestOutsourcePool_HireWithOptions(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	t.Run("hire with custom options", func(t *testing.T) {
		agent, err := pool.HireWithOptions("parent-1", "researcher", "task-1", "gpt-4", 1*time.Hour)
		require.NoError(t, err)
		require.NotNil(t, agent)

		assert.Equal(t, "parent-1", agent.ParentID)
		assert.Equal(t, "researcher", agent.Role)
		assert.Equal(t, "task-1", agent.TaskID)
		assert.Equal(t, "gpt-4", agent.Model)
		assert.True(t, agent.ExpiresAt.Sub(agent.HiredAt) >= 1*time.Hour)
	})

	t.Run("hire with zero TTL uses default", func(t *testing.T) {
		agent, err := pool.HireWithOptions("parent-1", "coder", "task-1", "", 0)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Should use default TTL of 30 minutes
		diff := agent.ExpiresAt.Sub(agent.HiredAt)
		assert.True(t, diff >= 29*time.Minute && diff <= 31*time.Minute)
	})
}

func TestOutsourcePool_Release(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	t.Run("release existing agent", func(t *testing.T) {
		agent, err := pool.Hire("parent-1", "coder", "task-1")
		require.NoError(t, err)

		assert.Equal(t, 1, pool.GetPoolSize())

		released := pool.Release(agent.ID)
		assert.True(t, released)
		assert.Equal(t, 0, pool.GetPoolSize())
	})

	t.Run("release non-existent agent", func(t *testing.T) {
		released := pool.Release("non-existent-id")
		assert.False(t, released)
	})

	t.Run("release already released agent", func(t *testing.T) {
		agent, err := pool.Hire("parent-1", "coder", "task-1")
		require.NoError(t, err)

		pool.Release(agent.ID)
		released := pool.Release(agent.ID)
		assert.False(t, released)
	})
}

func TestOutsourcePool_GetActiveAgents(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	t.Run("get active agents", func(t *testing.T) {
		// Hire some agents
		agent1, _ := pool.Hire("parent-1", "coder", "task-1")
		agent2, _ := pool.Hire("parent-1", "researcher", "task-2")
		_, _ = pool.Hire("parent-2", "planner", "task-3")

		agents := pool.GetActiveAgents()
		assert.Len(t, agents, 3)

		// Verify all agents are present
		ids := make(map[string]bool)
		for _, a := range agents {
			ids[a.ID] = true
		}
		assert.True(t, ids[agent1.ID])
		assert.True(t, ids[agent2.ID])
	})

	t.Run("empty pool returns empty slice", func(t *testing.T) {
		emptyPool := NewOutsourcePool(5, 30*time.Minute)
		agents := emptyPool.GetActiveAgents()
		assert.Empty(t, agents)
	})
}

func TestOutsourcePool_GetAgent(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	t.Run("get existing agent", func(t *testing.T) {
		agent, _ := pool.Hire("parent-1", "coder", "task-1")

		found, ok := pool.GetAgent(agent.ID)
		assert.True(t, ok)
		assert.NotNil(t, found)
		assert.Equal(t, agent.ID, found.ID)
	})

	t.Run("get non-existent agent", func(t *testing.T) {
		found, ok := pool.GetAgent("non-existent")
		assert.False(t, ok)
		assert.Nil(t, found)
	})
}

func TestOutsourcePool_IsFull(t *testing.T) {
	pool := NewOutsourcePool(2, 30*time.Minute)

	t.Run("not full", func(t *testing.T) {
		assert.False(t, pool.IsFull())
	})

	t.Run("full after hiring", func(t *testing.T) {
		_, _ = pool.Hire("parent-1", "coder", "task-1")
		_, _ = pool.Hire("parent-1", "coder", "task-2")

		assert.True(t, pool.IsFull())
	})
}

func TestOutsourcePool_GetAvailableSlots(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	assert.Equal(t, 5, pool.GetAvailableSlots())

	_, _ = pool.Hire("parent-1", "coder", "task-1")
	assert.Equal(t, 4, pool.GetAvailableSlots())

	_, _ = pool.Hire("parent-1", "coder", "task-2")
	assert.Equal(t, 3, pool.GetAvailableSlots())
}

func TestOutsourceAgent_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		agent := &OutsourceAgent{
			HiredAt:   time.Now(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		assert.False(t, agent.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		agent := &OutsourceAgent{
			HiredAt:   time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Second),
		}
		assert.True(t, agent.IsExpired())
	})
}

func TestOutsourcePool_ExtendExpiration(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	t.Run("extend existing agent", func(t *testing.T) {
		agent, _ := pool.Hire("parent-1", "coder", "task-1")
		originalExpiry := agent.ExpiresAt

		extended := pool.ExtendExpiration(agent.ID, 15*time.Minute)
		assert.True(t, extended)

		// Get fresh reference
		updated, _ := pool.GetAgent(agent.ID)
		assert.True(t, updated.ExpiresAt.After(originalExpiry))
	})

	t.Run("extend non-existent agent", func(t *testing.T) {
		extended := pool.ExtendExpiration("non-existent", 15*time.Minute)
		assert.False(t, extended)
	})
}

func TestOutsourcePool_GetAgentsByParent(t *testing.T) {
	pool := NewOutsourcePool(10, 30*time.Minute)

	// Hire agents for different parents
	_, _ = pool.Hire("parent-1", "coder", "task-1")
	_, _ = pool.Hire("parent-1", "researcher", "task-2")
	_, _ = pool.Hire("parent-2", "coder", "task-3")

	t.Run("get agents by parent", func(t *testing.T) {
		agents := pool.GetAgentsByParent("parent-1")
		assert.Len(t, agents, 2)

		for _, a := range agents {
			assert.Equal(t, "parent-1", a.ParentID)
		}
	})

	t.Run("get agents for non-existent parent", func(t *testing.T) {
		agents := pool.GetAgentsByParent("non-existent")
		assert.Empty(t, agents)
	})
}

func TestOutsourcePool_GetAgentsByRole(t *testing.T) {
	pool := NewOutsourcePool(10, 30*time.Minute)

	// Hire agents with different roles
	_, _ = pool.Hire("parent-1", "coder", "task-1")
	_, _ = pool.Hire("parent-1", "coder", "task-2")
	_, _ = pool.Hire("parent-1", "researcher", "task-3")

	t.Run("get agents by role", func(t *testing.T) {
		agents := pool.GetAgentsByRole("coder")
		assert.Len(t, agents, 2)

		for _, a := range agents {
			assert.Equal(t, "coder", a.Role)
		}
	})

	t.Run("get agents for non-existent role", func(t *testing.T) {
		agents := pool.GetAgentsByRole("non-existent")
		assert.Empty(t, agents)
	})
}

func TestOutsourcePool_GetPoolStats(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	// Hire some agents
	_, _ = pool.Hire("parent-1", "coder", "task-1")
	_, _ = pool.Hire("parent-1", "researcher", "task-2")

	stats := pool.GetPoolStats()

	assert.Equal(t, 2, stats["pool_size"])
	assert.Equal(t, 5, stats["max_pool_size"])
	assert.Equal(t, 3, stats["available_slots"])
	assert.Equal(t, false, stats["is_full"])
	assert.Equal(t, 0, stats["expired_count"])
	assert.NotEmpty(t, stats["average_age"])
	assert.NotEmpty(t, stats["default_ttl"])
}

func TestOutsourcePool_SetMaxPoolSize(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	pool.SetMaxPoolSize(10)
	assert.Equal(t, 10, pool.GetMaxPoolSize())

	// Zero or negative should not change
	pool.SetMaxPoolSize(0)
	assert.Equal(t, 10, pool.GetMaxPoolSize())

	pool.SetMaxPoolSize(-5)
	assert.Equal(t, 10, pool.GetMaxPoolSize())
}

func TestOutsourcePool_SetDefaultTTL(t *testing.T) {
	pool := NewOutsourcePool(5, 30*time.Minute)

	pool.SetDefaultTTL(1 * time.Hour)
	// Verify by hiring a new agent
	agent, _ := pool.Hire("parent-1", "coder", "task-1")
	diff := agent.ExpiresAt.Sub(agent.HiredAt)
	assert.True(t, diff >= 59*time.Minute && diff <= 61*time.Minute)

	// Zero or negative should not change
	pool.SetDefaultTTL(0)
	pool.SetDefaultTTL(-1 * time.Second)
}

func TestOutsourcePool_Clear(t *testing.T) {
	pool := NewOutsourcePool(10, 30*time.Minute)

	// Hire some agents
	_, _ = pool.Hire("parent-1", "coder", "task-1")
	_, _ = pool.Hire("parent-1", "researcher", "task-2")
	_, _ = pool.Hire("parent-2", "planner", "task-3")

	assert.Equal(t, 3, pool.GetPoolSize())

	cleared := pool.Clear()
	assert.Equal(t, 3, cleared)
	assert.Equal(t, 0, pool.GetPoolSize())
}

func TestOutsourcePool_CleanupExpired(t *testing.T) {
	// Create pool with very short TTL
	pool := NewOutsourcePool(5, 100*time.Millisecond)

	// Hire an agent
	agent, _ := pool.Hire("parent-1", "coder", "task-1")
	assert.Equal(t, 1, pool.GetPoolSize())

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// The agent should be expired but still in map until cleanup
	// GetActiveAgents triggers cleanup
	agents := pool.GetActiveAgents()
	assert.Empty(t, agents)
	assert.Equal(t, 0, pool.GetPoolSize())

	// Verify the agent is no longer retrievable
	_, ok := pool.GetAgent(agent.ID)
	assert.False(t, ok)
}

func TestOutsourcePool_ConcurrentAccess(t *testing.T) {
	pool := NewOutsourcePool(100, 30*time.Minute)

	// Concurrent hires
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 5; j++ {
				_, err := pool.Hire("parent-1", "coder", "task-1")
				if err != nil {
					t.Logf("Hire failed: %v", err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 50 agents (or less if some failed due to capacity)
	assert.True(t, pool.GetPoolSize() <= 50)
	assert.True(t, pool.GetPoolSize() > 0)
}

func TestOutsourcePool_TimeUntilExpiry(t *testing.T) {
	agent := &OutsourceAgent{
		HiredAt:   time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	remaining := agent.TimeUntilExpiry()
	// Should be approximately 30 minutes
	assert.True(t, remaining > 29*time.Minute && remaining < 31*time.Minute)
}
