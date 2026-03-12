# PicoClaw Agent Team - Testing Guide

## Quick Start

```bash
# Run all tests
./scripts/test_agent_team.sh

# Run specific Go tests
go test -v ./pkg/agent/...
go test -v ./pkg/mailbox/...

# Run frontend tests
cd ui && npm run typecheck && npm run build
```

---

## Backend Testing

### 1. Unit Tests

#### Mailbox System
```bash
# Test basic mailbox operations
go test -v ./pkg/mailbox/... -run TestMailbox

# Test priority queue
go test -v ./pkg/mailbox/... -run TestMailbox_Priority

# Test hub (multi-agent)
go test -v ./pkg/mailbox/... -run TestHub
```

#### Agent System
```bash
# Test agent lifecycle
go test -v ./pkg/agent/... -run TestAgentLoop_StartStop

# Test coordinator
go test -v ./pkg/agent/... -run TestJarvisCoordinator_StartStop

# Test task routing
go test -v ./pkg/agent/... -run TestRouteTask
```

### 2. Integration Test

Create a test file `cmd/test_agent_team/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "picoclaw/agent/pkg/agent"
    "picoclaw/agent/pkg/config"
    "picoclaw/agent/pkg/mailbox"
    "picoclaw/agent/pkg/memory"
)

func main() {
    // Setup
    mbHub := mailbox.NewHub(1000)
    memManager := memory.NewManager(memory.NewFileDB("/tmp/test_agents"))
    registry := agent.NewTeamRegistry("/tmp/test_agents", mbHub, memManager)
    
    // Initialize with default team
    if err := registry.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    // Create agent manager
    agentManager := agent.NewAgentManager(registry, mbHub, memManager)
    if err := agentManager.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    // Start coordinator
    coordinator := agent.NewJarvisCoordinator(registry, mbHub)
    if err := coordinator.Start(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ Agent Team started successfully!")
    fmt.Printf("Running agents: %v\n", agentManager.GetRunningAgents())
    
    // Send test task
    testTask := agent.TaskRequest{
        ID:          "test-task-1",
        Description: "Research about Go concurrency patterns",
        From:        "user",
    }
    
    taskJSON, _ := json.Marshal(testTask)
    coordinator.SendMessage("user", "jarvis", mailbox.MessageTypeTask, string(taskJSON), mailbox.PriorityNormal)
    
    fmt.Println("📨 Test task sent to Jarvis")
    
    // Wait and check
    time.Sleep(5 * time.Second)
    
    // Check mailbox
    if mb, ok := mbHub.Get("atlas"); ok {
        fmt.Printf("📬 Atlas mailbox: %d messages\n", mb.GetUnreadCount())
    }
    
    // Cleanup
    agentManager.StopAll()
    coordinator.Stop()
    
    fmt.Println("✅ Test completed!")
}
```

Run it:
```bash
go run cmd/test_agent_team/main.go
```

### 3. API Testing with curl

#### Start Server
```bash
# Build and run the server
go build -o picoclaw .
./picoclaw server
```

#### Test REST API
```bash
# Get agent list
curl http://localhost:8080/api/agents

# Get agent details
curl http://localhost:8080/api/agents/jarvis

# Get team stats
curl http://localhost:8080/api/agents/stats

# Update agent
curl -X PUT http://localhost:8080/api/agents/jarvis \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "status": "idle"}'
```

#### Test WebSocket
```bash
# Connect to WebSocket (using wscat)
npm install -g wscat
wscat -c ws://localhost:8080/ws

# Subscribe to events
{"type": "subscribe", "events": ["agent_moved", "agent_status_changed"]}

# Send chat message
{"type": "chat_message", "payload": {"content": "Hello agents!", "sessionId": "main"}}
```

---

## Frontend Testing

### 1. Development Server
```bash
cd ui
npm install
npm run dev

# Open http://localhost:5173
```

### 2. Manual Testing Checklist

#### Office Canvas
- [ ] Agents display with correct emoji avatars
- [ ] Status indicators show correct colors
- [ ] Click agent to select
- [ ] Animation states work (idle/working/talking)

#### Chat Sidebar
- [ ] Toggle sidebar with button
- [ ] Switch between Chat/Agents tabs
- [ ] Create new chat session (Main/Direct/Meeting)
- [ ] Send message triggers agent working animation
- [ ] Receive message triggers agent talking animation

#### Settings Modal
- [ ] Click agent settings icon
- [ ] Change model configuration
- [ ] Update persona settings
- [ ] Set schedule
- [ ] Change visual (emoji/color)

### 3. WebSocket Test Page

Create `ui/public/test.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <h1>Agent Team WebSocket Test</h1>
    <div id="status">Disconnected</div>
    <div id="messages"></div>
    
    <input type="text" id="input" placeholder="Type message...">
    <button onclick="send()">Send</button>
    
    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');
        const status = document.getElementById('status');
        const messages = document.getElementById('messages');
        
        ws.onopen = () => {
            status.textContent = 'Connected';
            status.style.color = 'green';
            
            // Subscribe to events
            ws.send(JSON.stringify({
                type: 'subscribe',
                events: ['chat_message', 'agent_status_changed', 'task_assigned']
            }));
        };
        
        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            const div = document.createElement('div');
            div.textContent = JSON.stringify(msg, null, 2);
            messages.appendChild(div);
        };
        
        ws.onclose = () => {
            status.textContent = 'Disconnected';
            status.style.color = 'red';
        };
        
        function send() {
            const input = document.getElementById('input');
            ws.send(JSON.stringify({
                type: 'chat_message',
                payload: {
                    content: input.value,
                    sessionId: 'main'
                }
            }));
            input.value = '';
        }
    </script>
</body>
</html>
```

---

## End-to-End Testing

### 1. Full System Test

```bash
# Terminal 1: Start backend
go run . server

# Terminal 2: Start frontend
cd ui && npm run dev

# Test flow:
# 1. Open http://localhost:5173
# 2. Verify agents appear on canvas
# 3. Open chat sidebar
# 4. Send message to Jarvis
# 5. Verify working animation
# 6. Wait for response
# 7. Verify talking animation
```

### 2. Load Testing

```bash
# Install hey (HTTP load tester)
go install github.com/rakyll/hey@latest

# Test REST API
hey -n 1000 -c 10 http://localhost:8080/api/agents

# Test WebSocket with custom script
```

### 3. Memory Testing

```bash
# Check memory usage while running
# In another terminal:
watch -n 1 'ps aux | grep picoclaw | grep -v grep'

# Or use go tool
go tool pprof http://localhost:8080/debug/pprof/heap
```

---

## Debugging

### Enable Debug Logs
```bash
# Set log level
export PICOCLAW_LOG_LEVEL=debug

# Run with verbose output
go run . server 2>&1 | tee server.log
```

### Check Agent Status
```bash
# Get stats
curl http://localhost:8080/api/agents/stats | jq

# Check specific agent
curl http://localhost:8080/api/agents/jarvis | jq
```

### WebSocket Debugging
```javascript
// In browser console
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (e) => console.log('Message:', JSON.parse(e.data));
ws.send(JSON.stringify({type: 'ping'}));
```

---

## CI/CD Testing

### GitHub Actions
```yaml
name: Test Agent Team

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Run Backend Tests
        run: go test -v ./pkg/agents/... ./pkg/mailbox/...
      
      - name: Build Backend
        run: go build .
      
      - name: Install UI Dependencies
        run: cd ui && npm ci
      
      - name: Test UI
        run: cd ui && npm run typecheck && npm run build
```

---

## Troubleshooting

### Common Issues

1. **Agent loop won't start**
   - Check if provider API key is set
   - Verify agent is enabled
   - Check logs for errors

2. **WebSocket not connecting**
   - Verify server is running
   - Check firewall settings
   - Try different port

3. **Animations not working**
   - Check browser console for errors
   - Verify PixiJS is loaded
   - Check agent status updates

4. **Memory issues**
   - Reduce mailbox capacity
   - Enable cleanup
   - Check for memory leaks

### Debug Commands
```bash
# Check running processes
ps aux | grep picoclaw

# Check network connections
netstat -tulpn | grep 8080

# View logs
tail -f ~/.picoclaw/logs/picoclaw.log

# Test config
cat ~/.picoclaw/config.json | jq
```
