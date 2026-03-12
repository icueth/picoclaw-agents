package heartbeat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"picoclaw/agent/pkg/tools"
)

func TestExecuteHeartbeat_Async(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{}) // Enable for testing

	asyncCalled := false
	asyncResult := &tools.ToolResult{
		ForLLM:  "Background task started",
		ForUser: "Task started in background",
		Silent:  false,
		IsError: false,
		Async:   true,
	}

	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		asyncCalled = true
		if prompt == "" {
			t.Error("Expected non-empty prompt")
		}
		return asyncResult
	})

	// Create HEARTBEAT.md
	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644)

	// Execute heartbeat directly (internal method for testing)
	hs.executeHeartbeat()

	if !asyncCalled {
		t.Error("Expected handler to be called")
	}
}

func TestExecuteHeartbeat_ResultLogging(t *testing.T) {
	tests := []struct {
		name    string
		result  *tools.ToolResult
		wantLog string
	}{
		{
			name: "error result",
			result: &tools.ToolResult{
				ForLLM:  "Heartbeat failed: connection error",
				ForUser: "",
				Silent:  false,
				IsError: true,
				Async:   false,
			},
			wantLog: "error message",
		},
		{
			name: "silent result",
			result: &tools.ToolResult{
				ForLLM:  "Heartbeat completed successfully",
				ForUser: "",
				Silent:  true,
				IsError: false,
				Async:   false,
			},
			wantLog: "completion message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			hs := NewHeartbeatService(tmpDir, 30, true)
			hs.stopChan = make(chan struct{}) // Enable for testing

			hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
				return tt.result
			})

			os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644)
			hs.executeHeartbeat()

			logFile := filepath.Join(tmpDir, "heartbeat.log")
			data, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}
			if string(data) == "" {
				t.Errorf("Expected log file to contain %s", tt.wantLog)
			}
		})
	}
}

func TestHeartbeatService_StartStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 1, true)

	err = hs.Start()
	if err != nil {
		t.Fatalf("Failed to start heartbeat service: %v", err)
	}

	hs.Stop()

	time.Sleep(100 * time.Millisecond)
}

func TestHeartbeatService_Disabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 1, false)

	if hs.enabled != false {
		t.Error("Expected service to be disabled")
	}

	err = hs.Start()
	_ = err // Disabled service returns nil
}

func TestExecuteHeartbeat_NilResult(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{}) // Enable for testing

	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		return nil
	})

	// Create HEARTBEAT.md
	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644)

	// Should not panic with nil result
	hs.executeHeartbeat()
}

// TestLogPath verifies heartbeat log is written to workspace directory
func TestLogPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)

	// Write a log entry
	hs.logf("INFO", "Test log entry")

	// Verify log file exists at workspace root
	expectedLogPath := filepath.Join(tmpDir, "heartbeat.log")
	if _, err := os.Stat(expectedLogPath); os.IsNotExist(err) {
		t.Errorf("Expected log file at %s, but it doesn't exist", expectedLogPath)
	}
}

// TestExtractUserTasks verifies task extraction from HEARTBEAT.md
func TestExtractUserTasks(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name:    "only template with separator",
			content: "# Heartbeat\n\nSome instructions\n\n---\n\nAdd your heartbeat tasks below this line:\n",
			want:    "",
		},
		{
			name:    "template with actual tasks",
			content: "# Heartbeat\n\n---\n\n- Check weather\n- Check email",
			want:    "- Check weather\n- Check email",
		},
		{
			name:    "no separator — backward compat",
			content: "- Check weather",
			want:    "- Check weather",
		},
		{
			name:    "separator with only whitespace after",
			content: "# Heartbeat\n---\n\n   \n\n",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUserTasks(tt.content)
			if got != tt.want {
				t.Errorf("extractUserTasks() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestExecuteHeartbeat_ConcurrencyGuard verifies overlapping heartbeats are skipped
func TestExecuteHeartbeat_ConcurrencyGuard(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{})

	callCount := 0
	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		callCount++
		time.Sleep(200 * time.Millisecond) // Simulate slow handler
		return tools.SilentResult("ok")
	})

	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	// Start first heartbeat in background
	go hs.executeHeartbeat()
	time.Sleep(50 * time.Millisecond) // Let it start

	// Second heartbeat should be skipped (concurrency guard)
	hs.executeHeartbeat()

	time.Sleep(300 * time.Millisecond) // Wait for first to finish

	if callCount != 1 {
		t.Errorf("Expected handler called once (guard should skip 2nd), got %d", callCount)
	}
}

// TestExecuteHeartbeat_ErrorBackoff verifies backoff after consecutive errors
func TestExecuteHeartbeat_ErrorBackoff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{})

	callCount := 0
	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		callCount++
		return &tools.ToolResult{IsError: true, ForLLM: "API error"}
	})

	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("- Check something"), 0o644)

	// Generate consecutive errors
	for i := 0; i < 6; i++ {
		hs.executeHeartbeat()
	}

	// Flow: call1=error(cnt=1), call2=error(cnt=2), call3=error(cnt=3),
	// call4=backoff(cnt→2), call5=error(cnt=3), call6=backoff(cnt→2)
	// Handler called on: 1, 2, 3, 5 = 4 times
	if callCount != 4 {
		t.Errorf("Expected 4 handler calls (backoff skips some), got %d", callCount)
	}
}

// TestBuildPrompt_SkipsEmptyTasks verifies no LLM call when HEARTBEAT.md has no tasks
func TestBuildPrompt_SkipsEmptyTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)

	// Write HEARTBEAT.md with only template (no actual tasks)
	template := `# Heartbeat Check List

This file contains tasks for the heartbeat service to check periodically.

---

Add your heartbeat tasks below this line:
`
	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte(template), 0o644)

	prompt := hs.buildPrompt()
	if prompt != "" {
		t.Errorf("Expected empty prompt for template-only HEARTBEAT.md, got: %s", prompt)
	}
}

// TestLogRotation verifies log file rotation when it exceeds max size
func TestLogRotation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	logPath := filepath.Join(tmpDir, "heartbeat.log")

	// Create a log file larger than maxLogBytes (256KB)
	bigLog := make([]byte, maxLogBytes+1024)
	for i := range bigLog {
		if i%80 == 79 {
			bigLog[i] = '\n'
		} else {
			bigLog[i] = 'x'
		}
	}
	os.WriteFile(logPath, bigLog, 0o644)

	hs.rotateLogIfNeeded()

	// Verify file was rotated (smaller than before)
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Log file missing after rotation: %v", err)
	}
	if info.Size() >= int64(len(bigLog)) {
		t.Errorf("Expected rotated log to be smaller, got %d bytes (was %d)", info.Size(), len(bigLog))
	}
}

// TestHeartbeatFilePath verifies HEARTBEAT.md is at workspace root
func TestHeartbeatFilePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)

	// Trigger default template creation
	hs.buildPrompt()

	// Verify HEARTBEAT.md exists at workspace root
	expectedPath := filepath.Join(tmpDir, "HEARTBEAT.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected HEARTBEAT.md at %s, but it doesn't exist", expectedPath)
	}
}
