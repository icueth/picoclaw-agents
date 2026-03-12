# PicoClaw Agents

## Multi-Agent System (A2A)

PicoClaw uses an 8-agent collaborative system:

### Core Agents

| Agent | Role | Avatar | Description |
|-------|------|--------|-------------|
| **Jarvis** | Coordinator | 🤖 | Main coordinator, manages and delegates tasks |
| **Atlas** | Researcher | 🕵️ | Information gathering and research |
| **Scribe** | Writer | ✍️ | Content creation and documentation |
| **Clawed** | Coder | 👨‍💻 | Code implementation and debugging |
| **Sentinel** | QA | 🛡️ | Quality assurance and testing |
| **Trendy** | Analyst | 📈 | Market trends and data analysis |
| **Pixel** | Designer | 🎨 | UI/UX and visual design |
| **Nova** | Architect | 🌌 | System architecture and planning |

## Usage

### Direct Agent Chat

```bash
picoclaw agent -m "Hello Jarvis"
picoclaw agent -a clawed -m "Write a Python script"
```

### Agent Delegation

Agents automatically delegate tasks based on capabilities:
- Research tasks → Atlas
- Coding tasks → Clawed
- Writing tasks → Scribe
- etc.

## Configuration

Each agent can be customized in `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "list": [
      {
        "id": "jarvis",
        "model": "kimi-for-coding",
        "capabilities": ["general", "coordination"]
      }
    ]
  }
}
```
