# PicoClaw Quick Start Guide (Enhanced Edition)

## Overview

PicoClaw is an ultra-lightweight Multi-Agent Orchestration system that runs on minimal hardware. This guide will get you up and running with your own AI Agent Team in minutes.

## Prerequisites

- **OS**: macOS, Linux, or Windows (WSL/Git Bash)
- **Go**: 1.21+ (for building from source)
- **Node.js**: 18+ (for Web UI)
- **Memory**: <10MB RAM for core binary, ~50MB for Web UI

## Installation

### Option 1: Pre-built Binary (Recommended)

Download the latest binary for your platform from [Releases](https://github.com/sipeed/picoclaw/releases).

```bash
# macOS (Apple Silicon)
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-darwin-arm64
chmod +x picoclaw-darwin-arm64
sudo mv picoclaw-darwin-arm64 /usr/local/bin/picoclaw

# macOS (Intel)
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-darwin-amd64
chmod +x picoclaw-darwin-amd64
sudo mv picoclaw-darwin-amd64 /usr/local/bin/picoclaw

# Linux (x86_64)
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-amd64
chmod +x picoclaw-linux-amd64
sudo mv picoclaw-linux-amd64 /usr/local/bin/picoclaw

# Linux (ARM64)
wget https://github.com/sipeed/picoclaw/releases/latest/download/picoclaw-linux-arm64
chmod +x picoclaw-linux-arm64
sudo mv picoclaw-linux-arm64 /usr/local/bin/picoclaw
```

### Option 2: Build from Source

```bash
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

# Build only
make build

# Build and install (auto-detects platform)
make install

# Or install to specific location
make install INSTALL_PREFIX=$HOME/.local
make install-user     # Install to ~/.local/bin
make install-system   # Install to /usr/local/bin (requires sudo)
```

**Check installation:**
```bash
make install-check    # Verify paths and environment
```

### Cross-Platform Notes

| Platform | Install Path | Notes |
|----------|--------------|-------|
| macOS | `~/.local/bin` or `/usr/local/bin` | Auto-adds to PATH |
| Linux | `~/.local/bin` or `/usr/local/bin` | Auto-detects shell (bash/zsh) |
| Windows | `%LOCALAPPDATA%\Programs` | Use Git Bash or WSL |

**Environment Variables:**
- `PICOCLAW_HOME`: Data directory (default: `~/.picoclaw`)
- `INSTALL_PREFIX`: Installation prefix
- `PICOCLAW_CONFIG`: Custom config file path

## Initial Setup

### 1. Initialize Configuration

```bash
# 1. Onboard core system
picoclaw onboard

# 2. Initialize Agent Team (Optional)
# This sets up the default Jarvis team structure
picoclaw team init --template generic
```

This creates:
- `~/.picoclaw/config.json` - Main configuration
- `~/.picoclaw/auth.json` - OAuth credentials
- `~/.picoclaw/workspace/` - Working directory with Agnet/User/Identity profiles
- `~/.picoclaw/picoclaw.db` - SQLite database for persistent memory and jobs

### 2. Configure API Keys

Edit `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model_name": "kimi-for-coding",
      "max_tokens": 32768,
      "restrict_to_workspace": false,
      "allow_read_outside_workspace": true
    }
  },
  "model_list": [
    {
      "model_name": "kimi-for-coding",
      "model": "kimi-coding/kimi-for-coding",
      "api_base": "https://api.kimi.com/coding/v1",
      "api_key": "YOUR_API_KEY"
    }
  ],
  "tools": {
    "exec": {
      "safety_level": "permissive"
    }
  }
}
```

**Default Security Settings:**
- `restrict_to_workspace: false` - Agents can access files outside workspace
- `allow_read_outside_workspace: true` - Allow reading external files
- `safety_level: "permissive"` - Permissive command execution

**Recommended API Providers:**
- [OpenRouter](https://openrouter.ai/keys) - Access 100+ models
- [Kimi (Moonshot)](https://platform.moonshot.cn/) - Good for coding
- [Anthropic](https://console.anthropic.com) - Claude models
- [OpenAI](https://platform.openai.com) - GPT models

### 3. Setup Embedding Service (Optional, for RAG)

```bash
# Setup Python embedding service
picoclaw setup embedding

# Or manually
make setup-embedding-user
```

## Usage

### Basic Chat

```bash
# Direct chat
picoclaw agent -m "What is 2+2?"

# Interactive mode
picoclaw agent

# Chat with specific agent
picoclaw agent -a jarvis -m "Hello!"
picoclaw agent -a clawed -m "Write a Python script"
```

### Multi-Agent System

PicoClaw includes 8 specialized agents:

| Agent | Role | Command |
|-------|------|---------|
| **Jarvis** | Coordinator (default) | `picoclaw agent -a jarvis` |
| **Atlas** | Researcher | `picoclaw agent -a atlas` |
| **Scribe** | Writer | `picoclaw agent -a scribe` |
| **Clawed** | Coder | `picoclaw agent -a clawed` |
| **Sentinel** | QA | `picoclaw agent -a sentinel` |
| **Trendy** | Analyst | `picoclaw agent -a trendy` |
| **Pixel** | Designer | `picoclaw agent -a pixel` |
| **Nova** | Architect | `picoclaw agent -a nova` |

Agents automatically delegate tasks based on capabilities.

### Gateway Mode (Chat Apps)

Run PicoClaw as a server for Telegram, Discord, etc.:

```bash
picoclaw gateway
```

Configure channels in `~/.picoclaw/config.json`:

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

See [Chat Apps section in README](../README.md#-chat-apps) for detailed setup.

### Scheduled Tasks (Cron)

```bash
# Add a scheduled task
picoclaw cron add "0 9 * * *" "Send daily report"

# List tasks
picoclaw cron list

# Remove task
picoclaw cron remove <id>
```

## File Locations

| File/Directory | Description |
|----------------|-------------|
| `~/.picoclaw/config.json` | Main configuration |
| `~/.picoclaw/workspace/` | Working directory |
| `~/.picoclaw/picoclaw.db` | SQLite database |
| `~/.picoclaw/models/` | Embedding models |
| `~/.picoclaw/agents/` | Per-agent data directories |

## Customizing Workspace Location

Set `PICOCLAW_HOME` environment variable:

```bash
# macOS/Linux
export PICOCLAW_HOME=/path/to/custom/location

# Windows (Git Bash)
export PICOCLAW_HOME=C:/Users/YourName/.picoclaw

# Then run
picoclaw onboard
```

## Troubleshooting

### "command not found: picoclaw"

```bash
# Check if binary is in PATH
which picoclaw

# If not, add to PATH manually
export PATH="$HOME/.local/bin:$PATH"

# Or reinstall with PATH setup
make install-user
```

### "config.json not found"

Run `picoclaw onboard` to initialize configuration.

### Embedding issues

If you are using OpenAI for embeddings, ensure your API key is correct. If using `local`, no external service is required.

### Permission denied on install

```bash
# Use user install (recommended)
make install-user

# Or fix permissions
sudo chown -R $(whoami) /usr/local/bin
```

## Next Steps

- Read [Tools Configuration](tools_configuration.md) for advanced tool setup
- Check [Troubleshooting](troubleshooting.md) for common issues
- See [README](../README.md) for full documentation
- Join [Discord](https://discord.gg/V4sAZ9XWpN) for community support
