# PicoClaw Agent Meeting System - Gateway Integration

## Overview

The Agent Meeting System is now fully integrated with the main PicoClaw gateway, providing HTTP API endpoints for managing meetings, scheduling, and AI-powered discussions.

## API Endpoints

### Health & Status
```
GET /health              - Health check with uptime
GET /ready               - Readiness check
```

### Agents
```
GET /api/agents              - List all 8 agents with basic info
GET /api/agents/{id}         - Get detailed agent info including persona
```

### Meetings
```
GET /api/meetings            - List all meetings
POST /api/meetings           - Create new meeting
GET /api/meetings/{id}       - Get meeting details with messages
POST /api/meetings/{id}      - Perform action (start, post, end)
DELETE /api/meetings/{id}    - Cancel meeting
```

### Scheduling
```
GET /api/schedule            - List scheduled meetings
POST /api/schedule           - Schedule new meeting
GET /api/schedule/upcoming   - Get upcoming meetings (24h default)
DELETE /api/schedule/{id}    - Cancel scheduled meeting
```

### AI Discussions
```
GET /api/discussions         - List AI discussions
POST /api/discussions        - Start AI-powered discussion
GET /api/discussions/{id}    - Get discussion with messages
```

## Usage Examples

### Create a Meeting
```bash
curl -X POST http://localhost:18790/api/meetings \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Feature Planning",
    "description": "Plan new feature implementation",
    "participants": ["atlas", "clawed", "nova"],
    "facilitator": "jarvis",
    "agenda": ["Requirements", "Design", "Timeline"],
    "auto_start": false
  }'
```

### Schedule a Meeting
```bash
curl -X POST http://localhost:18790/api/schedule \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Weekly Sync",
    "description": "Team synchronization",
    "scheduled_at": "2026-03-10T09:00:00Z",
    "participants": ["jarvis", "atlas", "clawed"],
    "facilitator": "jarvis",
    "reminder": "15m",
    "auto_start": true
  }'
```

### Start AI Discussion
```bash
curl -X POST http://localhost:18790/api/discussions \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "Architecture Review",
    "context": "Review microservices design",
    "participants": ["jarvis", "nova", "clawed"],
    "max_turns": 3
  }'
```

### Get Agent Details
```bash
curl http://localhost:18790/api/agents/jarvis
```

## File Structure

```
pkg/agent/
├── meeting/
│   ├── types.go            # Meeting, Participant, Message types
│   ├── conference.go       # ConferenceManager for meeting management
│   ├── scheduler.go        # Meeting scheduler with recurring support
│   ├── ai_discussion.go    # AI-powered agent discussions
│   └── api.go              # HTTP API handlers
├── persona/
│   ├── persona.go          # Persona file management
│   └── templates.go        # Default persona templates
```

## Agent Persona Files

Each agent has 3 persona files in `~/.picoclaw/agents/{agent_id}/`:

- **IDENTITY.md** - Role, responsibilities, origins
- **SOUL.md** - Personality, values, communication style
- **MEMORY.md** - Experience log, collaborations

## Components

### ConferenceManager
- Manages active meetings
- Facilitates agent participation
- Tracks meeting state and messages

### Scheduler
- Schedules meetings for future dates
- Supports recurring meetings (daily, weekly)
- Sends reminders before meetings
- Auto-starts meetings

### AI Discussion Manager
- Orchestrates AI-powered discussions
- Each agent responds based on their persona
- Generates consensus and summaries

### Meeting API Handler
- HTTP handlers for all endpoints
- JSON request/response format
- RESTful design

## Running the Gateway

```bash
# Build
make build

# Run gateway
./picoclaw gateway

# Or with debug
./picoclaw gateway --debug
```

## Testing

```bash
# Run test client
./test_gateway_api

# Or test manually
curl http://localhost:18790/api/agents
curl http://localhost:18790/api/meetings
```

## Integration Points

1. **Gateway** (`cmd/picoclaw/internal/gateway/helpers.go`)
   - Initializes meeting API on startup
   - Registers routes on channel manager's mux

2. **Channel Manager** (`pkg/channels/manager.go`)
   - Provides HTTP mux for route registration
   - Shared HTTP server with webhook handlers

3. **Agent System** (`pkg/agent/`)
   - Uses existing agent registry
   - Integrates with persona system
   - Works with LLM provider for AI discussions

## Future Enhancements

- WebSocket support for real-time updates
- Meeting recording and playback
- Advanced scheduling (calendar integration)
- Meeting analytics and insights
- Multi-language support for AI discussions
