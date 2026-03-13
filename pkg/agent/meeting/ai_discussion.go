// AI-powered discussion where agents respond using LLM
package meeting

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/agent/persona"
	"picoclaw/agent/pkg/providers"
)

// AIDiscussionManager manages AI-powered agent discussions
type AIDiscussionManager struct {
	provider      providers.LLMProvider
	conference    *ConferenceManager
	personaBaseDir string
	
	// Active AI discussions
	mu           sync.RWMutex
	discussions  map[string]*AIDiscussion
}

// AIDiscussion represents an ongoing AI-powered discussion
type AIDiscussion struct {
	ID           string
	Topic        string
	Context      string
	Participants []AIAgentParticipant
	Messages     []AIDiscussionMessage
	Status       string // "ongoing", "completed", "paused"
	MaxTurns     int
	CurrentTurn  int
	Facilitator  string
	Consensus    string // Final consensus from discussion
	
	// Control
	discCtx    context.Context
	cancel     context.CancelFunc
	onMessage  func(AIDiscussionMessage)
	onComplete func(string) // consensus
}

// AIAgentParticipant represents an agent in AI discussion
type AIAgentParticipant struct {
	AgentID    string
	Name       string
	Avatar     string
	Role       string
	PersonaDir string
	SystemPrompt string // Loaded from persona files
}

// AIDiscussionMessage represents a message in AI discussion
type AIDiscussionMessage struct {
	Turn      int
	AgentID   string
	Name      string
	Avatar    string
	Content   string
	Type      string // "statement", "question", "proposal", "agreement", "objection"
	Timestamp time.Time
}

// NewAIDiscussionManager creates a new AI discussion manager
func NewAIDiscussionManager(provider providers.LLMProvider, conference *ConferenceManager, personaBaseDir string) *AIDiscussionManager {
	return &AIDiscussionManager{
		provider:       provider,
		conference:     conference,
		personaBaseDir: personaBaseDir,
		discussions:    make(map[string]*AIDiscussion),
	}
}

// StartAIDiscussion begins an AI-powered discussion
func (dm *AIDiscussionManager) StartAIDiscussion(parentCtx context.Context, topic, discussionContext string, agentIDs []string, maxTurns int) (*AIDiscussion, error) {
	if dm.provider == nil {
		return nil, fmt.Errorf("LLM provider not available")
	}

	// Load participants with their personas
	participants := make([]AIAgentParticipant, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		agentInfo, ok := dm.conference.GetAgent(agentID)
		if !ok {
			continue
		}

		personaDir := dm.personaBaseDir + "/" + agentID
		systemPrompt := dm.buildSystemPrompt(agentInfo, personaDir, topic)

		participants = append(participants, AIAgentParticipant{
			AgentID:      agentID,
			Name:         agentInfo.Name,
			Avatar:       agentInfo.Avatar,
			Role:         agentInfo.Role,
			PersonaDir:   personaDir,
			SystemPrompt: systemPrompt,
		})
	}

	if len(participants) == 0 {
		return nil, fmt.Errorf("no valid participants")
	}

	discCtx, cancel := context.WithCancel(parentCtx)
	
	discussion := &AIDiscussion{
		ID:           generateAIDiscussionID(),
		Topic:        topic,
		Context:      discussionContext,
		Participants: participants,
		Messages:     make([]AIDiscussionMessage, 0),
		Status:       "ongoing",
		MaxTurns:     maxTurns,
		CurrentTurn:  0,
		Facilitator:  "jarvis",
		discCtx:      discCtx,
		cancel:       cancel,
	}

	dm.mu.Lock()
	dm.discussions[discussion.ID] = discussion
	dm.mu.Unlock()

	// Start the AI discussion in background
	go dm.runAIDiscussion(discussion)

	return discussion, nil
}

// runAIDiscussion orchestrates the AI discussion
func (dm *AIDiscussionManager) runAIDiscussion(disc *AIDiscussion) {
	defer func() {
		disc.Status = "completed"
		if disc.onComplete != nil {
			disc.onComplete(dm.generateConsensus(disc))
		}
	}()

	// Opening statement from facilitator
	opening := dm.generateOpeningStatement(disc)
	dm.addMessage(disc, disc.Facilitator, "Jarvis", "🤖", opening, "statement")

	// Round-robin discussion
	for turn := 0; turn < disc.MaxTurns && disc.Status == "ongoing"; turn++ {
		disc.CurrentTurn = turn
		
		for i, participant := range disc.Participants {
			// Skip facilitator (Jarvis) in regular turns - he only facilitates
			if participant.AgentID == disc.Facilitator {
				continue
			}

			// Generate response for this agent
			response, msgType := dm.generateAgentResponse(disc, participant, turn, i)
			
			dm.addMessage(disc, participant.AgentID, participant.Name, participant.Avatar, response, msgType)
			
			// Small delay to simulate thinking
			select {
			case <-time.After(500 * time.Millisecond):
			case <-disc.discCtx.Done():
				return
			}
		}

		// Facilitator summarizes every few turns
		if turn > 0 && turn%3 == 0 {
			summary := dm.generateFacilitatorSummary(disc)
			dm.addMessage(disc, disc.Facilitator, "Jarvis", "🤖", summary, "summary")
		}
	}

	// Final summary and consensus
	finalSummary := dm.generateFinalSummary(disc)
	dm.addMessage(disc, disc.Facilitator, "Jarvis", "🤖", finalSummary, "summary")
	
	// Set consensus
	disc.Consensus = dm.generateConsensus(disc)
}

// generateAgentResponse generates a response using LLM
func (dm *AIDiscussionManager) generateAgentResponse(disc *AIDiscussion, participant AIAgentParticipant, turn, index int) (string, string) {
	// Build conversation history
	history := dm.buildConversationHistory(disc, 5) // Last 5 messages

	// Build prompt
	prompt := fmt.Sprintf(`%s

## Current Discussion
**Topic**: %s
**Context**: %s
**Your Turn**: %d

## Conversation So Far
%s

## Your Role in This Discussion
As %s (%s), respond to the discussion. Consider:
- Your expertise and perspective
- What other agents have said
- Moving the discussion toward a solution
- Ask questions if you need clarification
- Propose specific ideas or solutions

Respond in character as %s. Be concise (2-4 sentences) but insightful.`,
		participant.SystemPrompt,
		disc.Topic,
		disc.Context,
		turn+1,
		history,
		participant.Name,
		participant.Role,
		participant.Name,
	)

	// Call LLM
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	response, err := dm.callLLM(ctx, prompt)
	if err != nil {
		// Fallback response
		return fmt.Sprintf("As %s, I think we should consider the technical feasibility of these options.", participant.Name), "statement"
	}

	// Determine message type
	msgType := dm.classifyMessageType(response)

	return response, msgType
}

// buildSystemPrompt creates system prompt from persona files
func (dm *AIDiscussionManager) buildSystemPrompt(agentInfo *AgentInfo, personaDir, topic string) string {
	var parts []string

	// Load persona summary
	personaSummary, _ := persona.GetPersonaSummary(personaDir)
	if personaSummary != "" {
		parts = append(parts, "## Your Identity\n"+personaSummary)
	}

	// Add role-specific context
	parts = append(parts, fmt.Sprintf("\n## Current Context\nYou are participating in a team discussion about: %s", topic))
	parts = append(parts, "\n## Communication Guidelines\n- Stay in character based on your identity and soul")
	parts = append(parts, "- Reference your past experiences when relevant")
	parts = append(parts, "- Collaborate with other agents respectfully")
	parts = append(parts, "- Be specific and actionable in your suggestions")

	return strings.Join(parts, "\n")
}

// buildConversationHistory formats recent messages for context
func (dm *AIDiscussionManager) buildConversationHistory(disc *AIDiscussion, limit int) string {
	if len(disc.Messages) == 0 {
		return "(Just starting)"
	}

	start := len(disc.Messages) - limit
	if start < 0 {
		start = 0
	}

	var history strings.Builder
	for _, msg := range disc.Messages[start:] {
		history.WriteString(fmt.Sprintf("%s %s: %s\n", msg.Avatar, msg.Name, msg.Content))
	}

	return history.String()
}

// generateOpeningStatement creates facilitator's opening
func (dm *AIDiscussionManager) generateOpeningStatement(disc *AIDiscussion) string {
	participants := make([]string, 0)
	for _, p := range disc.Participants {
		if p.AgentID != disc.Facilitator {
			participants = append(participants, fmt.Sprintf("%s %s (%s)", p.Avatar, p.Name, p.Role))
		}
	}

	return fmt.Sprintf("👋 Welcome everyone! Today we're discussing: **%s**\n\n📋 Context: %s\n\n👥 Participants: %s\n\n💡 Let's collaborate and find the best solution. Each of you will have a chance to share your perspective. Feel free to ask questions, propose ideas, or build on what others say.",
		disc.Topic,
		disc.Context,
		strings.Join(participants, ", "),
	)
}

// generateFacilitatorSummary creates a summary of recent discussion
func (dm *AIDiscussionManager) generateFacilitatorSummary(disc *AIDiscussion) string {
	// Get recent messages
	recent := disc.Messages
	if len(recent) > 6 {
		recent = recent[len(recent)-6:]
	}

	var points []string
	for _, msg := range recent {
		if msg.Type == "proposal" || msg.Type == "agreement" {
			points = append(points, fmt.Sprintf("- %s suggested: %s", msg.Name, msg.Content))
		}
	}

	if len(points) == 0 {
		return "📝 So far we've heard various perspectives. Let's continue exploring our options."
	}

	return "📝 **Summary of key points so far:**\n" + strings.Join(points, "\n") + "\n\nLet's continue building on these ideas."
}

// generateFinalSummary creates the final meeting summary
func (dm *AIDiscussionManager) generateFinalSummary(disc *AIDiscussion) string {
	var proposals []string
	var agreements []string

	for _, msg := range disc.Messages {
		if msg.Type == "proposal" {
			proposals = append(proposals, msg.Content)
		} else if msg.Type == "agreement" {
			agreements = append(agreements, msg.Content)
		}
	}

	summary := "✅ **Discussion Complete**\n\n"
	summary += "📊 **Summary:**\n"
	
	if len(proposals) > 0 {
		summary += "\n**Key Proposals:**\n"
		for i, p := range proposals {
			if i < 3 { // Top 3
				summary += fmt.Sprintf("%d. %s\n", i+1, p)
			}
		}
	}

	if len(agreements) > 0 {
		summary += "\n**Points of Agreement:**\n"
		for _, a := range agreements {
			summary += "- " + a + "\n"
		}
	}

	consensus := dm.generateConsensus(disc)
	summary += fmt.Sprintf("\n🤝 **Consensus:** %s\n", consensus)
	summary += "\nThank you all for your valuable contributions! 🎉"

	return summary
}

// generateConsensus extracts consensus from discussion
func (dm *AIDiscussionManager) generateConsensus(disc *AIDiscussion) string {
	// Look for agreements and common themes
	agreementCount := 0
	for _, msg := range disc.Messages {
		if msg.Type == "agreement" || msg.Type == "proposal" {
			agreementCount++
		}
	}

	if agreementCount >= 3 {
		return "Team reached consensus on proposed approach with strong agreement on key points."
	} else if agreementCount > 0 {
		return "Team found partial agreement; further discussion needed on specific details."
	}
	return "Team shared perspectives but requires additional discussion to reach full consensus."
}

// classifyMessageType determines the type of message
func (dm *AIDiscussionManager) classifyMessageType(content string) string {
	contentLower := strings.ToLower(content)
	
	if strings.Contains(contentLower, "?") {
		return "question"
	}
	if strings.Contains(contentLower, "propose") || strings.Contains(contentLower, "suggest") || strings.Contains(contentLower, "we should") {
		return "proposal"
	}
	if strings.Contains(contentLower, "agree") || strings.Contains(contentLower, "support") {
		return "agreement"
	}
	if strings.Contains(contentLower, "disagree") || strings.Contains(contentLower, "concern") || strings.Contains(contentLower, "but") {
		return "objection"
	}
	return "statement"
}

// addMessage adds a message to the discussion
func (dm *AIDiscussionManager) addMessage(disc *AIDiscussion, agentID, name, avatar, content, msgType string) {
	msg := AIDiscussionMessage{
		Turn:      disc.CurrentTurn,
		AgentID:   agentID,
		Name:      name,
		Avatar:    avatar,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now(),
	}

	disc.Messages = append(disc.Messages, msg)

	if disc.onMessage != nil {
		disc.onMessage(msg)
	}
}

// callLLM makes the actual LLM call
func (dm *AIDiscussionManager) callLLM(callCtx context.Context, prompt string) (string, error) {
	if dm.provider == nil {
		return "", fmt.Errorf("no provider")
	}

	// Create a simple chat completion request
	messages := []providers.Message{
		{Role: "system", Content: "You are an AI agent participating in a team discussion. Respond concisely in character."},
		{Role: "user", Content: prompt},
	}

	response, err := dm.provider.Chat(callCtx, messages, nil, "", nil)
	if err != nil {
		return "", err
	}

	if response == nil || response.Content == "" {
		return "", fmt.Errorf("empty response")
	}

	return response.Content, nil
}

// GetDiscussion returns an AI discussion by ID
func (dm *AIDiscussionManager) GetDiscussion(id string) (*AIDiscussion, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	disc, ok := dm.discussions[id]
	return disc, ok
}

// SetCallbacks sets event callbacks for a discussion
func (disc *AIDiscussion) SetCallbacks(onMessage func(AIDiscussionMessage), onComplete func(string)) {
	disc.onMessage = onMessage
	disc.onComplete = onComplete
}

// Stop stops the AI discussion
func (disc *AIDiscussion) Stop() {
	disc.cancel()
	disc.Status = "stopped"
}

// Helper
func generateAIDiscussionID() string {
	return fmt.Sprintf("ai-disc-%d", time.Now().UnixNano())
}
