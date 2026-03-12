# Troubleshooting

## Installation Issues

### "make: command not found"

**Cause:** Make is not installed on your system.

**Fix:**
```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt-get install make

# Fedora/RHEL
sudo dnf install make

# Windows (Git Bash)
# Make is included with Git for Windows
```

### "command not found: picoclaw" after installation

**Cause:** The installation directory is not in your PATH.

**Fix:**
```bash
# Check where picoclaw was installed
which picoclaw || find ~/.local -name picoclaw 2>/dev/null

# Add to PATH (temporary)
export PATH="$HOME/.local/bin:$PATH"

# Add to PATH (permanent - bash)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Add to PATH (permanent - zsh)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Reinstall with auto PATH setup
make install-user
```

### Wrong PICOCLAW_HOME path on Windows

**Cause:** Windows uses different path formats than Unix.

**Fix:**
```bash
# Git Bash
export PICOCLAW_HOME="C:/Users/YourName/.picoclaw"

# WSL
export PICOCLAW_HOME="/mnt/c/Users/YourName/.picoclaw"

# Or use Windows-style with forward slashes
export PICOCLAW_HOME="$USERPROFILE/.picoclaw"
```

### Permission denied during install

**Cause:** Insufficient permissions for system directories.

**Fix:**
```bash
# Option 1: Use user install (recommended)
make install-user

# Option 2: Fix permissions
sudo chown -R $(whoami) /usr/local/bin

# Option 3: Use sudo (not recommended for development)
sudo make install-system
```

## Configuration Issues

### "config.json not found"

**Cause:** `picoclaw onboard` hasn't been run.

**Fix:**
```bash
# Initialize configuration
picoclaw onboard

# Or with custom location
PICOCLAW_HOME=/custom/path picoclaw onboard
```

### Config changes not taking effect

**Cause:** PicoClaw caches config or you edited the wrong file.

**Fix:**
```bash
# Find the correct config file
picoclaw status  # Shows config path

# Or check directly
cat ~/.picoclaw/config.json

# Restart gateway after config changes
pkill picoclaw
picoclaw gateway
```

### "workspace not found"

**Cause:** Workspace path is misconfigured or not created.

**Fix:**
```bash
# Check workspace path in config
picoclaw agent -m "show my workspace path"

# Create manually if needed
mkdir -p ~/.picoclaw/workspace

# Or re-run onboard
picoclaw onboard
```

## Model/Provider Issues

### "model ... not found in model_list"

**Symptom:** You see either:
- `Error creating provider: model "openrouter/free" not found in model_list`
- OpenRouter returns 400: `"free is not a valid model ID"`

**Cause:** The `model` field in your `model_list` entry is what gets sent to the API. For OpenRouter you must use the **full** model ID, not a shorthand.

- **Wrong:** `"model": "free"` → OpenRouter receives `free` and rejects it.
- **Right:** `"model": "openrouter/free"` → OpenRouter receives `openrouter/free` (auto free-tier routing).

**Fix:** In `~/.picoclaw/config.json`:

1. **agents.defaults.model** must match a `model_name` in `model_list` (e.g., `"openrouter-free"`).
2. That entry's **model** must be a valid OpenRouter model ID, for example:
   - `"openrouter/free"` – auto free-tier
   - `"google/gemini-2.0-flash-exp:free"`
   - `"meta-llama/llama-3.1-8b-instruct:free"`

Example snippet:

```json
{
  "agents": {
    "defaults": {
      "model": "openrouter-free"
    }
  },
  "model_list": [
    {
      "model_name": "openrouter-free",
      "model": "openrouter/free",
      "api_key": "sk-or-v1-YOUR_OPENROUTER_KEY",
      "api_base": "https://openrouter.ai/api/v1"
    }
  ]
}
```

Get your key at [OpenRouter Keys](https://openrouter.ai/keys).

### "API key not found"

**Cause:** API key is not set in config or environment variable.

**Fix:**
```bash
# Edit config
vim ~/.picoclaw/config.json

# Or use environment variable
export PICOCLAW_AGENTS_DEFAULTS_MODEL=gpt4
export PICOCLAW_MODEL_LIST_0_API_KEY=your-key-here
```


## Gateway/Channel Issues

### "Address already in use"

**Cause:** Another process is using the gateway port (default: 18790).

**Fix:**
```bash
# Find and kill the process
lsof -i :18790
kill <PID>

# Or use a different port
# Edit ~/.picoclaw/config.json:
# "gateway": { "port": 18791 }
```

### Webhook channels not receiving messages

**Cause:** Gateway is bound to localhost (127.0.0.1) which is not accessible externally.

**Fix:**
```json
{
  "gateway": {
    "host": "0.0.0.0",
    "port": 18790
  }
}
```

**Security Note:** Only use `0.0.0.0` if you have proper firewall rules or are behind a reverse proxy.

### Telegram bot not responding

**Cause:** Webhook not set or incorrect token.

**Fix:**
```bash
# Check bot token
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe

# For development, delete webhook to use polling
picoclaw gateway  # Uses polling mode by default
```

## Runtime Issues

### "Permission denied" when executing commands

**Cause:** Safety level is blocking the command.

**Fix:**
```json
{
  "tools": {
    "exec": {
      "safety_level": "permissive",
      "enable_deny_patterns": true
    }
  }
}
```

Or for specific commands:
```json
{
  "tools": {
    "exec": {
      "custom_allow_patterns": ["\\bdocker\\s+run\\b"]
    }
  }
}
```

### High memory usage

**Cause:** Multiple agents running simultaneously or large context windows.

**Fix:**
```bash
# Check running processes
picoclaw status

# Kill and restart
pkill picoclaw
picoclaw gateway

# Reduce max_tokens in config
# "max_tokens": 4096  # instead of 32768
```

### Agent not responding

**Cause:** Agent loop may be stuck or provider is slow.

**Fix:**
```bash
# Check logs
tail -f ~/.picoclaw/logs/picoclaw.log

# Restart gateway
pkill picoclaw
picoclaw gateway --debug
```

## Debug Mode

Enable debug logging for detailed output:

```bash
# Debug mode
picoclaw gateway --debug

# Or set environment variable
export PICOCLAW_LOG_LEVEL=debug
picoclaw gateway
```

## Agent Team Issues

### "Agent [name] failed to start"

**Cause:** Persona file missing or model configuration invalid.

**Fix:**
```bash
# Verify agent configuration
picoclaw team status

# Check for required files in workspace
ls ~/.picoclaw/workspace/IDENTITY.md
ls ~/.picoclaw/workspace/AGENT_[NAME].md

# Test agent in isolation
picoclaw agent --id [agent_id] -m "ping"
```

### "Mailbox delivery failed"

**Cause:** The agent's mailbox is full or the mailbox hub is disconnected.

**Fix:**
```bash
# Check mailbox status
picoclaw mailing list

# Restart the hub
pkill picoclaw
picoclaw gateway
```

### "RAG query returning empty results"

**Cause:** Vector database not indexed or embedding model is not configured correctly.

**Fix:**
```bash
# Verify config (model_name/model should be valid)
cat ~/.picoclaw/config.json

# Re-index documents
picoclaw rag index /path/to/docs

# Test simple query
picoclaw rag query "What is PicoClaw?"
```

### "Project state not saved"

**Cause:** SQLite database is locked or disk space is full.

**Fix:**
```bash
# Check database integrity
sqlite3 ~/.picoclaw/picoclaw.db "PRAGMA integrity_check;"

# Clear temp files
rm -rf /tmp/picoclaw_*
```

## Getting Help

If issues persist:

1. Check [GitHub Issues](https://github.com/sipeed/picoclaw/issues)
2. Join [Discord Community](https://discord.gg/V4sAZ9XWpN)
3. Run diagnostics:
   ```bash
   make install-check
   picoclaw status
   picoclaw version
   ```
