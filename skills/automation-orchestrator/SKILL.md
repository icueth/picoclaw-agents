---
name: automation-orchestrator
description: Intelligent workflow automation and orchestration system for AI agents with scheduling, error handling, and monitoring capabilities
---

# Automation Orchestrator

This built-in skill provides intelligent workflow automation and orchestration capabilities for AI agents to create, manage, and execute complex automated workflows with robust error handling and monitoring.

## Capabilities

- **Workflow Design**: Create complex workflows with conditional logic, loops, and parallel execution
- **Task Scheduling**: Schedule workflows to run at specific times, intervals, or event triggers
- **Error Handling**: Implement comprehensive error handling with retries, fallbacks, and notifications
- **State Management**: Track workflow state and progress with persistent storage
- **Resource Coordination**: Coordinate resource usage across multiple workflow steps
- **Monitoring and Logging**: Monitor workflow execution with detailed logging and metrics
- **Dependency Management**: Manage dependencies between workflow steps and external systems
- **Parameter Passing**: Pass data between workflow steps with type validation
- **Version Control**: Track workflow versions and enable rollback capabilities
- **Integration Hub**: Integrate with external services, APIs, and notification systems

## Usage Examples

### Simple Workflow
```yaml
tool: automation-orchestrator
action: create_workflow
workflow:
  name: "Daily Backup"
  description: "Daily backup of important files"
  steps:
    - name: "Check disk space"
      tool: "system-monitor"
      action: "get_system_status"
      parameters:
        metrics: ["disk_usage"]
    - name: "Create backup"
      tool: "file-backup"
      action: "create_backup"
      parameters:
        source: ["/home/user/documents"]
        destination: "/backup/daily"
      condition: "disk_usage < 80"
    - name: "Send notification"
      tool: "agent-mail"
      action: "send_email"
      parameters:
        to: ["user@example.com"]
        subject: "Daily Backup Completed"
        body: "Backup completed successfully"
```

### Scheduled Workflow
```yaml
tool: automation-orchestrator
action: schedule_workflow
workflow_id: "weekly-report"
schedule:
  type: "cron"
  expression: "0 9 * * 1"  # Every Monday at 9 AM
parameters:
  report_type: "weekly"
  recipients: ["team@example.com"]
error_handling:
  max_retries: 3
  retry_delay: "300s"
  notify_on_failure: true
```

### Complex Conditional Workflow
```yaml
tool: automation-orchestrator
action: execute_workflow
workflow:
  name: "Deployment Pipeline"
  steps:
    - name: "Run tests"
      tool: "secure-command-executor"
      action: "execute"
      parameters:
        command: "npm test"
    - name: "Build application"
      tool: "secure-command-executor"
      action: "execute"
      parameters:
        command: "npm run build"
      condition: "previous_step.success"
    - name: "Deploy to staging"
      tool: "docker-management"
      action: "deploy_container"
      parameters:
        image: "app:latest"
        environment: "staging"
      condition: "previous_step.success"
    - name: "Run integration tests"
      tool: "secure-command-executor"
      action: "execute"
      parameters:
        command: "npm run integration-tests"
      condition: "previous_step.success"
    - name: "Deploy to production"
      tool: "docker-management"
      action: "deploy_container"
      parameters:
        image: "app:latest"
        environment: "production"
      condition: "previous_step.success"
error_handling:
  on_failure:
    - name: "Rollback deployment"
      tool: "docker-management"
      action: "rollback_deployment"
    - name: "Notify team"
      tool: "agent-chat"
      action: "send_message"
      parameters:
        platform: "slack"
        channel: "#deployments"
        message: "Deployment failed, rollback initiated"
```

## Security Considerations

- Workflow definitions are validated against security policies before execution
- Sensitive parameters and credentials are encrypted and securely managed
- Access control ensures only authorized agents can create or modify workflows
- Audit logging tracks all workflow activities for compliance and security
- Resource limits prevent workflows from consuming excessive system resources

## Configuration

The automation-orchestrator skill can be configured with the following parameters:

- `default_timeout`: Default timeout for workflow steps (default: 5m)
- `max_concurrent_workflows`: Maximum number of concurrent workflows (default: 10)
- `retry_policy`: Default retry policy for failed steps (exponential_backoff)
- `storage_backend`: Storage backend for workflow state (sqlite, postgres, memory)
- `notification_channels`: Enabled notification channels (email, slack, internal)

This skill is essential for any agent that needs to automate complex workflows, coordinate multiple tools and services, handle errors gracefully, and ensure reliable execution of automated processes.