package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/fileutil"
	"picoclaw/agent/pkg/logger"
)

// DepartmentMemory provides shared knowledge storage for all agents in a department.
// This allows agents to share learnings, best practices, and common knowledge.
type DepartmentMemory struct {
	mu           sync.RWMutex
	department   string
	basePath     string
	knowledgeDir string
	insightsFile string
	bestPractices []BestPractice
	insights     []DepartmentInsight
}

// BestPractice represents a documented best practice for the department
type BestPractice struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Content     string    `json:"content"`
	CreatedBy   string    `json:"created_by"`   // Agent ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UseCount    int       `json:"use_count"`
	Tags        []string  `json:"tags"`
}

// DepartmentInsight represents a learned insight or pattern discovered by agents
type DepartmentInsight struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`       // Agent ID that discovered this
	Confidence  float64   `json:"confidence"`   // 0.0 to 1.0
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
	UseCount    int       `json:"use_count"`
	RelatedTo   []string  `json:"related_to"`   // IDs of related insights
}

// DepartmentKnowledge represents a knowledge document
type DepartmentKnowledge struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Source   string `json:"source"`
	Created  time.Time `json:"created"`
}

// NewDepartmentMemory creates or loads a department's shared memory
func NewDepartmentMemory(workspaceRoot, department string) *DepartmentMemory {
	if department == "" {
		department = "general"
	}

	basePath := filepath.Join(workspaceRoot, "agents", department)
	dm := &DepartmentMemory{
		department:   department,
		basePath:     basePath,
		knowledgeDir: filepath.Join(basePath, "knowledge"),
		insightsFile: filepath.Join(basePath, "insights.json"),
		bestPractices: make([]BestPractice, 0),
		insights:     make([]DepartmentInsight, 0),
	}

	// Ensure directories exist
	os.MkdirAll(dm.knowledgeDir, 0o755)
	
	// Load existing data
	dm.loadInsights()
	dm.loadBestPractices()

	return dm
}

// ==================== Best Practices ====================

// AddBestPractice adds a new best practice to the department
func (dm *DepartmentMemory) AddBestPractice(bp BestPractice) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if bp.ID == "" {
		bp.ID = fmt.Sprintf("bp-%d", time.Now().UnixNano())
	}
	bp.CreatedAt = time.Now()
	bp.UpdatedAt = bp.CreatedAt

	dm.bestPractices = append(dm.bestPractices, bp)
	
	logger.InfoCF("dept_memory", "Added best practice",
		map[string]any{
			"department": dm.department,
			"practice":   bp.Title,
			"by":         bp.CreatedBy,
		})

	return dm.saveBestPractices()
}

// GetBestPractices returns all best practices, optionally filtered by category
func (dm *DepartmentMemory) GetBestPractices(category string) []BestPractice {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if category == "" {
		// Return copy
		result := make([]BestPractice, len(dm.bestPractices))
		copy(result, dm.bestPractices)
		return result
	}

	var filtered []BestPractice
	for _, bp := range dm.bestPractices {
		if strings.EqualFold(bp.Category, category) {
			filtered = append(filtered, bp)
		}
	}
	return filtered
}

// FindBestPractices searches best practices by keywords
func (dm *DepartmentMemory) FindBestPractices(query string) []BestPractice {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []BestPractice

	for _, bp := range dm.bestPractices {
		if strings.Contains(strings.ToLower(bp.Title), query) ||
		   strings.Contains(strings.ToLower(bp.Content), query) ||
		   containsAnyTag(bp.Tags, query) {
			matches = append(matches, bp)
		}
	}

	return matches
}

// RecordPracticeUse increments the use count for a best practice
func (dm *DepartmentMemory) RecordPracticeUse(practiceID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for i := range dm.bestPractices {
		if dm.bestPractices[i].ID == practiceID {
			dm.bestPractices[i].UseCount++
			dm.bestPractices[i].UpdatedAt = time.Now()
			dm.saveBestPractices()
			return
		}
	}
}

// ==================== Insights ====================

// AddInsight adds a new department insight
func (dm *DepartmentMemory) AddInsight(insight DepartmentInsight) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if insight.ID == "" {
		insight.ID = fmt.Sprintf("in-%d", time.Now().UnixNano())
	}
	insight.CreatedAt = time.Now()
	insight.LastUsed = insight.CreatedAt

	dm.insights = append(dm.insights, insight)
	
	logger.InfoCF("dept_memory", "Added insight",
		map[string]any{
			"department": dm.department,
			"category":   insight.Category,
			"by":         insight.Source,
		})

	return dm.saveInsights()
}

// GetInsights returns insights, optionally filtered by category and confidence threshold
func (dm *DepartmentMemory) GetInsights(category string, minConfidence float64) []DepartmentInsight {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var filtered []DepartmentInsight
	for _, in := range dm.insights {
		if in.Confidence >= minConfidence {
			if category == "" || strings.EqualFold(in.Category, category) {
				filtered = append(filtered, in)
			}
		}
	}
	return filtered
}

// FindRelevantInsights searches for insights relevant to a query
func (dm *DepartmentMemory) FindRelevantInsights(query string, maxResults int) []DepartmentInsight {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	query = strings.ToLower(query)
	
	// Simple keyword matching with scoring
	type scoredInsight struct {
		insight DepartmentInsight
		score   int
	}
	
	var scored []scoredInsight
	
	for _, in := range dm.insights {
		score := 0
		content := strings.ToLower(in.Content)
		category := strings.ToLower(in.Category)
		
		// Score based on keyword matches
		keywords := strings.Fields(query)
		for _, kw := range keywords {
			if strings.Contains(content, kw) {
				score += 2
			}
			if strings.Contains(category, kw) {
				score += 3
			}
		}
		
		// Boost by confidence and use count
		score += int(in.Confidence * 5)
		score += in.UseCount
		
		if score > 0 {
			scored = append(scored, scoredInsight{in, score})
		}
	}
	
	// Sort by score (simple bubble sort for small lists)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	
	// Return top results
	if len(scored) > maxResults {
		scored = scored[:maxResults]
	}
	
	result := make([]DepartmentInsight, len(scored))
	for i, s := range scored {
		result[i] = s.insight
	}
	return result
}

// RecordInsightUse updates last used time and increments count
func (dm *DepartmentMemory) RecordInsightUse(insightID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for i := range dm.insights {
		if dm.insights[i].ID == insightID {
			dm.insights[i].UseCount++
			dm.insights[i].LastUsed = time.Now()
			dm.saveInsights()
			return
		}
	}
}

// ==================== Knowledge Documents ====================

// AddKnowledgeDocument adds a knowledge document to the department
func (dm *DepartmentMemory) AddKnowledgeDocument(doc DepartmentKnowledge) error {
	if doc.ID == "" {
		doc.ID = fmt.Sprintf("kn-%d", time.Now().UnixNano())
	}
	doc.Created = time.Now()

	filename := fmt.Sprintf("%s_%s.json", sanitizeFilename(doc.Category), doc.ID)
	filepath := filepath.Join(dm.knowledgeDir, filename)

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return fileutil.WriteFileAtomic(filepath, data, 0o644)
}

// GetKnowledgeDocuments returns all knowledge documents for the department
func (dm *DepartmentMemory) GetKnowledgeDocuments(category string) ([]DepartmentKnowledge, error) {
	entries, err := os.ReadDir(dm.knowledgeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []DepartmentKnowledge{}, nil
		}
		return nil, err
	}

	var docs []DepartmentKnowledge
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Filter by category if specified
		if category != "" && !strings.HasPrefix(entry.Name(), sanitizeFilename(category)+"_") {
			continue
		}

		filepath := filepath.Join(dm.knowledgeDir, entry.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			continue
		}

		var doc DepartmentKnowledge
		if err := json.Unmarshal(data, &doc); err == nil {
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

// ==================== Persistence ====================

func (dm *DepartmentMemory) loadInsights() {
	data, err := os.ReadFile(dm.insightsFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &dm.insights)
}

func (dm *DepartmentMemory) saveInsights() error {
	data, err := json.MarshalIndent(dm.insights, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(dm.insightsFile, data, 0o644)
}

func (dm *DepartmentMemory) loadBestPractices() {
	bpFile := filepath.Join(dm.basePath, "best_practices.json")
	data, err := os.ReadFile(bpFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &dm.bestPractices)
}

func (dm *DepartmentMemory) saveBestPractices() error {
	bpFile := filepath.Join(dm.basePath, "best_practices.json")
	data, err := json.MarshalIndent(dm.bestPractices, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(bpFile, data, 0o644)
}

// ==================== Context Generation ====================

// BuildContext generates a context string with department knowledge
// This can be injected into agent prompts
func (dm *DepartmentMemory) BuildContext(topic string) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var sb strings.Builder

	// Add relevant best practices
	practices := dm.FindBestPractices(topic)
	if len(practices) > 0 {
		sb.WriteString("## Department Best Practices\n\n")
		for i, bp := range practices {
			if i >= 3 { // Limit to top 3
				break
			}
			sb.WriteString(fmt.Sprintf("### %s\n", bp.Title))
			sb.WriteString(bp.Content)
			sb.WriteString("\n\n")
		}
	}

	// Add relevant insights
	insights := dm.FindRelevantInsights(topic, 3)
	if len(insights) > 0 {
		sb.WriteString("## Department Insights\n\n")
		for i, in := range insights {
			if i >= 3 { // Limit to top 3
				break
			}
			sb.WriteString(fmt.Sprintf("- [%s, confidence: %.0f%%] %s\n", 
				in.Category, in.Confidence*100, in.Content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ==================== Helper Functions ====================

func containsAnyTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func sanitizeFilename(name string) string {
	// Replace unsafe characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(strings.ToLower(name))
}

// ListAllDepartments returns all departments that have shared memory
func ListAllDepartments(workspaceRoot string) ([]string, error) {
	agentsDir := filepath.Join(workspaceRoot, "agents")
	
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var departments []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this is a department (has insights.json or best_practices.json)
			deptPath := filepath.Join(agentsDir, entry.Name())
			if hasDepartmentMemory(deptPath) {
				departments = append(departments, entry.Name())
			}
		}
	}

	return departments, nil
}

func hasDepartmentMemory(deptPath string) bool {
	insightsFile := filepath.Join(deptPath, "insights.json")
	bpFile := filepath.Join(deptPath, "best_practices.json")
	knowledgeDir := filepath.Join(deptPath, "knowledge")

	_, hasInsights := os.Stat(insightsFile)
	_, hasBP := os.Stat(bpFile)
	_, hasKnowledge := os.Stat(knowledgeDir)

	return hasInsights == nil || hasBP == nil || hasKnowledge == nil
}
