package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordScorer_Match(t *testing.T) {
	scorer := NewKeywordScorer()

	agentTiers := []AgentTier{
		{AgentID: "coder", Capabilities: []string{"code", "debug", "refactor"}},
		{AgentID: "researcher", Capabilities: []string{"research", "analysis", "information"}},
		{AgentID: "writer", Capabilities: []string{"writing", "creative", "content"}},
	}

	tests := []struct {
		name           string
		task           string
		expectedAgent  string
		minConfidence  float64
		shouldMatch    bool
	}{
		// Code-related tasks
		{
			name:          "Simple code request",
			task:          "Write a Python function to sort a list",
			expectedAgent: "coder",
			minConfidence: 0.7,
			shouldMatch:   true,
		},
		{
			name:          "Debug request",
			task:          "Help me debug this error in my JavaScript code",
			expectedAgent: "coder",
			minConfidence: 0.7,
			shouldMatch:   true,
		},
		{
			name:          "Thai code request",
			task:          "เขียนโค้ด Python ให้หน่อย",
			expectedAgent: "coder",
			minConfidence: 0.7,
			shouldMatch:   true,
		},
		{
			name:          "Thai debug request",
			task:          "แก้บักในโค้ดนี้ที",
			expectedAgent: "coder",
			minConfidence: 0.6,
			shouldMatch:   true,
		},

		// Research-related tasks
		{
			name:          "Research request",
			task:          "Find information about climate change",
			expectedAgent: "researcher",
			minConfidence: 0.6,
			shouldMatch:   true,
		},
		{
			name:          "Analysis request",
			task:          "Analyze the latest market trends",
			expectedAgent: "researcher",
			minConfidence: 0.6,
			shouldMatch:   true,
		},
		{
			name:          "Thai research request",
			task:          "หาข้อมูลเกี่ยวกับ AI",
			expectedAgent: "researcher",
			minConfidence: 0.7,
			shouldMatch:   true,
		},
		{
			name:          "Thai analysis request",
			task:          "วิเคราะห์ข่าวล่าสุด",
			expectedAgent: "researcher",
			minConfidence: 0.6,
			shouldMatch:   true,
		},

		// Writing-related tasks
		{
			name:          "Writing request",
			task:          "Write an article about technology",
			expectedAgent: "writer",
			minConfidence: 0.7,
			shouldMatch:   true,
		},
		{
			name:          "Creative writing",
			task:          "Create a short story about space exploration",
			expectedAgent: "writer",
			minConfidence: 0.7,
			shouldMatch:   true,
		},
		{
			name:          "Thai writing request",
			task:          "เขียนบทความเรื่องเทคโนโลยี",
			expectedAgent: "writer",
			minConfidence: 0.8,
			shouldMatch:   true,
		},
		{
			name:          "Thai story request",
			task:          "เขียนเรื่องสั้นเกี่ยวกับอวกาศ",
			expectedAgent: "writer",
			minConfidence: 0.8,
			shouldMatch:   true,
		},

		// Ambiguous tasks - should still match but with lower confidence
		{
			name:          "Ambiguous write request",
			task:          "Write something for me",
			expectedAgent: "writer",
			minConfidence: 0.5,
			shouldMatch:   true,
		},

		// Negative test - greeting should not strongly match any agent
		{
			name:          "Greeting only",
			task:          "Hello, how are you?",
			expectedAgent: "",
			minConfidence: 0.0,
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.Match(tt.task, agentTiers)

			if !tt.shouldMatch {
				// For non-matching cases, result might be nil or have low confidence
				if result != nil && result.Confidence > 0.5 {
					t.Errorf("Expected no strong match, got %s with confidence %.2f", result.AgentID, result.Confidence)
				}
				return
			}

			assert.NotNil(t, result, "Expected a match but got nil")
			if result != nil {
				assert.Equal(t, tt.expectedAgent, result.AgentID, "Expected agent %s but got %s", tt.expectedAgent, result.AgentID)
				assert.GreaterOrEqual(t, result.Confidence, tt.minConfidence,
					"Expected confidence >= %.2f but got %.2f", tt.minConfidence, result.Confidence)
				t.Logf("Matched '%s' with confidence %.2f (hits: %d)", result.AgentID, result.Confidence, result.Hits)
			}
		})
	}
}

func TestKeywordScorer_WeightedKeywords(t *testing.T) {
	scorer := NewKeywordScorer()

	agentTiers := []AgentTier{
		{AgentID: "coder", Capabilities: []string{"code", "debug"}},
	}

	// Test that primary keywords have higher weight than secondary
	tasks := []struct {
		task        string
		description string
	}{
		{"code", "single primary keyword"},
		{"programming debug", "two primary keywords"},
		{"script api", "two secondary keywords"},
		{"code script", "primary + secondary"},
	}

	for _, tc := range tasks {
		result := scorer.Match(tc.task, agentTiers)
		if result != nil {
			t.Logf("Task '%s': matched with confidence %.2f, hits: %d (%s)",
				tc.task, result.Confidence, result.Hits, tc.description)
		}
	}
}

func TestIsThai(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"Hello world", false},
		{"สวัสดี", true},
		{"Hello สวัสดี", true},
		{"Python programming", false},
		{"เขียนโค้ด", true},
		{"12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := IsThai(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
		delta    float64
	}{
		{
			name:     "Identical vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0, 0},
			expected: 1.0,
			delta:    0.0001,
		},
		{
			name:     "Orthogonal vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{0, 1, 0},
			expected: 0.0,
			delta:    0.0001,
		},
		{
			name:     "Opposite vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{-1, 0, 0},
			expected: -1.0,
			delta:    0.0001,
		},
		{
			name:     "45 degree angle",
			a:        []float32{1, 0},
			b:        []float32{1, 1},
			expected: 0.7071, // 1/sqrt(2)
			delta:    0.0001,
		},
		{
			name:     "Different lengths",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0},
			expected: 0.0, // Should return 0 for different lengths
			delta:    0.0001,
		},
		{
			name:     "Zero vector",
			a:        []float32{0, 0, 0},
			b:        []float32{1, 0, 0},
			expected: 0.0,
			delta:    0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

func BenchmarkKeywordScorer_Match(b *testing.B) {
	scorer := NewKeywordScorer()
	agentTiers := []AgentTier{
		{AgentID: "coder", Capabilities: []string{"code", "debug", "refactor"}},
		{AgentID: "researcher", Capabilities: []string{"research", "analysis", "information"}},
		{AgentID: "writer", Capabilities: []string{"writing", "creative", "content"}},
	}

	tasks := []string{
		"Write a Python function to calculate fibonacci numbers",
		"Find the latest news about artificial intelligence",
		"Create a blog post about healthy eating habits",
		"Debug this error in my JavaScript code",
		"เขียนโค้ด Python ให้หน่อย",
		"หาข้อมูลเกี่ยวกับ machine learning",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := tasks[i%len(tasks)]
		scorer.Match(task, agentTiers)
	}
}
