---
name: agent-task-manager
description: Comprehensive task management system for AI agents with state tracking and workflow orchestration
---

# Agent Task Manager

This built-in skill provides a robust task management system for AI agents to handle complex workflows, track task states, manage dependencies, and coordinate multi-step processes.

## Capabilities

- **Task Creation**: Create tasks with descriptions, priorities, deadlines, and metadata
- **State Management**: Track task states (pending, in_progress, completed, failed, blocked)
- **Dependencies**: Define task dependencies and execution order
- **Progress Tracking**: Monitor task progress with percentage completion and status updates
- **Error Handling**: Handle task failures with retry mechanisms and error reporting
- **Notifications**: Send notifications for task completion, failures, or deadlines
- **Task Groups**: Organize related tasks into groups or projects
- **Resource Allocation**: Manage resource allocation and concurrency limits
- **History Logging**: Maintain complete task execution history for audit trails
- **Export/Import**: Export task data to various formats and import from external sources

## Usage Examples

### Basic Task Creation
```yaml
tool: agent-task-manager
action: create_task
task:
  name: "Deploy Application"
  description: "Deploy the latest version to production"
  priority: "high"
  deadline: "2026-03-15T18:00:00Z"
  tags: ["deployment", "production"]
```

### Task Dependencies
```yaml
tool: agent-task-manager
action: create_task_group
group:
  name: "Feature Implementation"
  tasks:
    - name: "Design API"
      dependencies: []
    - name: "Implement Backend"
      dependencies: ["Design API"]
    - name: "Implement Frontend"
      dependencies: ["Design API"]
    - name: "Integration Testing"
      dependencies: ["Implement Backend", "Implement Frontend"]
```

### Progress Tracking
```yaml
tool: agent-task-manager
action: update_task
task_id: "task_12345"
updates:
  status: "in_progress"
  progress: 75
  notes: "Backend implementation completed, working on frontend integration"
```

## Security Considerations

- Task data is encrypted at rest when sensitive information is involved
- Access control ensures only authorized agents can modify tasks
- Audit logging tracks all task modifications for security compliance
- Resource limits prevent runaway task creation or execution

## Configuration

The agent-task-manager skill can be configured with the following parameters:

- `storage_backend`: Storage backend (sqlite, postgres, memory)
- `max_concurrent_tasks`: Maximum number of concurrent tasks (default: 10)
- `retry_policy`: Retry policy for failed tasks (max_retries, backoff_strategy)
- `notification_channels`: Notification channels (email, webhook, internal)
- `retention_policy`: Data retention policy for completed tasks

This skill is essential for any agent that needs to manage complex workflows, coordinate multiple steps, or track progress across extended operations.