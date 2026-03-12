# PicoClaw Workspace

This is your PicoClaw workspace directory where agents store and manage files.

## Directory Structure

```
workspace/
├── README.md           # This file
├── AGENTS.md          # Agent-specific documentation
├── skills/            # Installed skills directory
├── cron/              # Scheduled jobs data
└── [agent-folders]/   # Per-agent working directories
```

## Configuration

Main config file: `~/.picoclaw/config.json`

### Default Settings

- **Workspace Restriction**: `restrict_to_workspace: false` - Agents can access files outside workspace
- **Read Outside Workspace**: `allow_read_outside_workspace: true` - Allow reading external files
- **Safety Level**: `safety_level: "permissive"` - Permissive execution mode

## Quick Start

1. Edit `~/.picoclaw/config.json` to add your API keys
2. Start the gateway: `picoclaw gateway`
3. Chat with agents: `picoclaw agent -m "Hello!"`

## Security Notes

With `safety_level: "permissive"`, agents can:
- Execute most shell commands
- Read files outside workspace
- Write to allowed paths

Review `~/.picoclaw/config.json` to adjust security settings.
