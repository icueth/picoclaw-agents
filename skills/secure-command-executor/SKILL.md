---
name: secure-command-executor
description: Secure command execution system for AI agents with sandboxing, validation, and monitoring capabilities
---

# Secure Command Executor

This built-in skill provides secure command execution capabilities for AI agents to run system commands, scripts, and utilities while maintaining security, isolation, and auditability.

## Capabilities

- **Command Validation**: Validate commands against allowlists, denylists, and security policies
- **Sandboxed Execution**: Run commands in isolated environments with restricted permissions
- **Resource Limiting**: Enforce CPU, memory, disk, and network usage limits
- **Timeout Management**: Automatically terminate long-running or hung processes
- **Input Sanitization**: Sanitize command inputs to prevent injection attacks
- **Output Capture**: Capture and process command output (stdout, stderr, exit codes)
- **Real-time Monitoring**: Monitor command execution in real-time with progress tracking
- **Audit Logging**: Log all command executions with user context and parameters
- **Environment Control**: Manage execution environment variables and working directories
- **Error Handling**: Provide detailed error reporting and recovery mechanisms

## Usage Examples

### Execute Simple Command
```yaml
tool: secure-command-executor
action: execute
command: "ls -la /home/user/documents"
working_directory: "/tmp"
timeout: "30s"
resource_limits:
  cpu_cores: 1
  memory_mb: 100
  disk_mb: 50
capture_output: true
```

### Run Script with Parameters
```yaml
tool: secure-command-executor
action: execute_script
script_path: "/scripts/backup.sh"
parameters:
  - "--source=/home/user"
  - "--destination=/backup"
  - "--compress"
environment:
  BACKUP_KEY: "{{backup_encryption_key}}"
timeout: "10m"
sandbox_enabled: true
```

### Batch Command Execution
```yaml
tool: secure-command-executor
action: batch_execute
commands:
  - "git status"
  - "git log --oneline -5"
  - "npm list --depth=0"
working_directory: "/project"
concurrent: false
timeout_per_command: "15s"
```

### Monitor Long-running Process
```yaml
tool: secure-command-executor
action: monitor_process
command: "python training_script.py --epochs=100"
working_directory: "/ml-project"
timeout: "2h"
resource_limits:
  cpu_cores: 4
  memory_gb: 8
  gpu_enabled: true
progress_pattern: "Epoch (\\d+)/100"
```

## Security Considerations

- All commands are validated against security policies before execution
- Sandbox environments prevent access to sensitive system resources
- Input sanitization prevents command injection and shell injection attacks
- Resource limits prevent denial-of-service through resource exhaustion
- Audit logging provides complete traceability for security investigations
- Environment isolation prevents interference between concurrent executions

## Configuration

The secure-command-executor skill can be configured with the following parameters:

- `default_timeout`: Default command timeout (default: 30s)
- `max_timeout`: Maximum allowed timeout (default: 10m)
- `sandbox_enabled`: Enable sandboxed execution by default (default: true)
- `allowed_commands`: Allowlist of permitted commands (default: comprehensive safe list)
- `resource_limits`: Default resource limits for all executions
- `audit_level`: Audit logging level (minimal, standard, comprehensive)

This skill is essential for any agent that needs to execute system commands, run scripts, automate workflows, or interact with the underlying operating system while maintaining security and reliability.