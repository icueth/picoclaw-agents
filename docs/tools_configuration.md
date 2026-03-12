# Tools Configuration

PicoClaw's tools configuration is located in the `tools` field of `config.json`.

## Directory Structure

```json
{
  "tools": {
    "web": { ... },
    "mcp": { ... },
    "exec": { ... },
    "cron": { ... },
    "skills": { ... },
    "subagent": { ... },
    "rag": { ... },
    "memory": { ... },
    "project": { ... }
  }
}
```

## Web Tools

Web tools are used for web search and fetching.

### Brave

| Config        | Type   | Default | Description               |
| ------------- | ------ | ------- | ------------------------- |
| `enabled`     | bool   | false   | Enable Brave search       |
| `api_key`     | string | -       | Brave Search API key      |
| `max_results` | int    | 5       | Maximum number of results |

### DuckDuckGo

| Config        | Type | Default | Description               |
| ------------- | ---- | ------- | ------------------------- |
| `enabled`     | bool | true    | Enable DuckDuckGo search  |
| `max_results` | int  | 5       | Maximum number of results |

### Perplexity

| Config        | Type   | Default | Description               |
| ------------- | ------ | ------- | ------------------------- |
| `enabled`     | bool   | false   | Enable Perplexity search  |
| `api_key`     | string | -       | Perplexity API key        |
| `max_results` | int    | 5       | Maximum number of results |

## Exec Tool

The exec tool is used to execute shell commands.

| Config                 | Type   | Default     | Description                                |
| ---------------------- | ------ | ----------- | ------------------------------------------ |
| `safety_level`         | string | "permissive"| Safety level: "strict", "balanced", or "permissive" |
| `enable_deny_patterns` | bool   | true        | Enable default dangerous command blocking  |
| `custom_deny_patterns` | array  | []          | Custom deny patterns (regular expressions) |
| `custom_allow_patterns`| array  | []          | Custom allow patterns (override blocks)    |

### Safety Levels

The `safety_level` setting controls which command patterns are blocked:

- **`strict`** - Blocks all potentially dangerous patterns (most restrictive)
  - Includes: file deletion, disk operations, system commands, privilege escalation
  - Use for: Production environments, shared systems
  
- **`balanced`** (formerly default) - Blocks critical and cautious patterns
  - Includes: `rm -rf`, `format`, `dd`, `curl | sh`, etc.
  - Allows: Most development commands
  - Use for: Development workstations
  
- **`permissive`** (default since v0.2+) - Blocks only critical patterns
  - Blocks: `rm -rf /`, disk wipes, fork bombs, `curl | bash`
  - Allows: Most commands including `sudo`, `docker`, package managers
  - Use for: Personal development, containerized environments

### Functionality

- **`safety_level`**: Choose from "strict", "balanced", or "permissive"
- **`enable_deny_patterns`**: Set to `false` to completely disable pattern blocking
- **`custom_deny_patterns`**: Add custom deny regex patterns
- **`custom_allow_patterns`**: Add patterns that override denies (whitelist)

### Default Blocked Command Patterns by Level

**Permissive (Default):**
- Critical deletes: `rm -rf /`, `rm -rf /*`
- Disk wipes: `dd if=* of=/dev/sd*`, `mkfs`, `format`
- Fork bombs: `:(){ :|:& };:`
- Pipe to shell: `curl * | sh`, `wget * | sh`

**Balanced (Additional to Permissive):**
- Delete commands: `rm -rf`, `del /f/q`, `rmdir /s`
- System operations: `shutdown`, `reboot`, `poweroff`
- Privilege escalation: `sudo`, `chmod`, `chown`
- Process control: `pkill`, `killall`, `kill -9`
- Package management: `apt`, `yum`, `dnf`, `npm install -g`
- Containers: `docker run`, `docker exec`

**Strict (Additional to Balanced):**
- Command substitution: `$()`, `${}`, backticks
- Remote operations: `ssh`, `scp`, `rsync`
- Git operations: `git push`, `git force`
- Script execution: `eval`, `source *.sh`

### Configuration Examples

**Permissive mode (default, recommended for development):**
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

**Balanced mode (recommended for team environments):**
```json
{
  "tools": {
    "exec": {
      "safety_level": "balanced",
      "enable_deny_patterns": true,
      "custom_allow_patterns": ["\\bdocker\\s+run\\b"]
    }
  }
}
```

**Strict mode (recommended for production):**
```json
{
  "tools": {
    "exec": {
      "safety_level": "strict",
      "enable_deny_patterns": true,
      "custom_deny_patterns": ["\\brm\\s+-r\\b", "\\bkillall\\s+python"]
    }
  }
}
```

## Cron Tool

The cron tool is used for scheduling periodic tasks.

| Config                 | Type | Default | Description                                    |
| ---------------------- | ---- | ------- | ---------------------------------------------- |
| `exec_timeout_minutes` | int  | 5       | Execution timeout in minutes, 0 means no limit |

## MCP Tool

The MCP tool enables integration with external Model Context Protocol servers.

### Global Config

| Config    | Type   | Default | Description                         |
| --------- | ------ | ------- | ----------------------------------- |
| `enabled` | bool   | false   | Enable MCP integration globally     |
| `servers` | object | `{}`    | Map of server name to server config |

### Per-Server Config

| Config     | Type   | Required | Description                                |
| ---------- | ------ | -------- | ------------------------------------------ |
| `enabled`  | bool   | yes      | Enable this MCP server                     |
| `type`     | string | no       | Transport type: `stdio`, `sse`, `http`     |
| `command`  | string | stdio    | Executable command for stdio transport     |
| `args`     | array  | no       | Command arguments for stdio transport      |
| `env`      | object | no       | Environment variables for stdio process    |
| `env_file` | string | no       | Path to environment file for stdio process |
| `url`      | string | sse/http | Endpoint URL for `sse`/`http` transport    |
| `headers`  | object | no       | HTTP headers for `sse`/`http` transport    |

### Transport Behavior

- If `type` is omitted, transport is auto-detected:
  - `url` is set → `sse`
  - `command` is set → `stdio`
- `http` and `sse` both use `url` + optional `headers`.
- `env` and `env_file` are only applied to `stdio` servers.

### Configuration Examples

#### 1) Stdio MCP server

```json
{
  "tools": {
    "mcp": {
      "enabled": true,
      "servers": {
        "filesystem": {
          "enabled": true,
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
        }
      }
    }
  }
}
```

#### 2) Remote SSE/HTTP MCP server

```json
{
  "tools": {
    "mcp": {
      "enabled": true,
      "servers": {
        "remote-mcp": {
          "enabled": true,
          "type": "sse",
          "url": "https://example.com/mcp",
          "headers": {
            "Authorization": "Bearer YOUR_TOKEN"
          }
        }
      }
    }
  }
}
```

## Skills Tool

The skills tool configures skill discovery and installation via registries like ClawHub.

### Registries

| Config                             | Type   | Default              | Description             |
| ---------------------------------- | ------ | -------------------- | ----------------------- |
| `registries.clawhub.enabled`       | bool   | true                 | Enable ClawHub registry |
| `registries.clawhub.base_url`      | string | `https://clawhub.ai` | ClawHub base URL        |
| `registries.clawhub.search_path`   | string | `/api/v1/search`     | Search API path         |
| `registries.clawhub.skills_path`   | string | `/api/v1/skills`     | Skills API path         |
| `registries.clawhub.download_path` | string | `/api/v1/download`   | Download API path       |

### Configuration Example

```json
{
  "tools": {
    "skills": {
      "registries": {
        "clawhub": {
          "enabled": true,
          "base_url": "https://clawhub.ai",
          "search_path": "/api/v1/search",
          "skills_path": "/api/v1/skills",
          "download_path": "/api/v1/download"
        }
      }
    }
  }
}
```

## Agent Team Tools

These tools enable the Multi-Agent Orchestration features of the Enhanced Edition.

### Subagent System

The subagent tool allows an agent to spawn specialized sub-tasks.

| Config | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | true | Enable subagent spawning |
| `max_depth` | int | 2 | Maximum recursion depth for subagents |
| `default_role` | string | "assistant" | Default role for new subagents |
| `allow_parallel` | bool | true | Allow multiple subagents to run concurrently |

### RAG (Retrieval-Augmented Generation)

Tools for storing and querying vector-based knowledge.

| Config | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | true | Enable RAG tools |
| `embedding_model` | string | "local" | Model for embeddings (local/openai/http). "local" is Go-native (built-in). |
| `vector_store` | string | "sqlite" | Storage backend for vectors |
| `chunk_size` | int | 1000 | Text chunking size (characters) |

### Persistent Memory

Enhanced memory tools for cross-session recall.

| Config | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | true | Enable persistent memory tools |
| `db_path` | string | "~/.picoclaw/picoclaw.db" | Path to SQLite database |
| `max_recall_results` | int | 5 | Max results for semantic memory recall |

### Project Management

Tools for A2A project coordination and state management.

| Config | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | true | Enable project management tools |
| `auto_save` | bool | true | Automatically save project state on changes |

## Environment Variables

All configuration options can be overridden via environment variables with the format `PICOCLAW_TOOLS_<SECTION>_<KEY>`:

For example:

- `PICOCLAW_TOOLS_WEB_BRAVE_ENABLED=true`
- `PICOCLAW_TOOLS_EXEC_ENABLE_DENY_PATTERNS=false`
- `PICOCLAW_TOOLS_CRON_EXEC_TIMEOUT_MINUTES=10`
- `PICOCLAW_TOOLS_MCP_ENABLED=true`

Note: Nested map-style config (for example `tools.mcp.servers.<name>.*`) is configured in `config.json` rather than environment variables.
