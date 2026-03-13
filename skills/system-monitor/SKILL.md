---
name: system-monitor
description: Comprehensive system monitoring and diagnostics for AI agents with real-time metrics and alerting capabilities
---

# System Monitor

This built-in skill provides comprehensive system monitoring and diagnostics capabilities for AI agents to track system health, performance metrics, and resource utilization in real-time.

## Capabilities

- **Resource Monitoring**: Track CPU, memory, disk, network, and GPU usage in real-time
- **Process Management**: Monitor and manage running processes with detailed information
- **System Health Checks**: Perform comprehensive system health assessments and diagnostics
- **Performance Metrics**: Collect and analyze performance metrics with historical trends
- **Alert Generation**: Generate alerts for threshold violations and system anomalies
- **Log Analysis**: Analyze system logs for errors, warnings, and performance issues
- **Network Monitoring**: Monitor network connections, bandwidth usage, and latency
- **Disk Usage Analysis**: Track disk space usage, I/O performance, and file system health
- **Temperature Monitoring**: Monitor system temperatures and thermal throttling events
- **Custom Metrics**: Define and track custom system metrics and KPIs

## Usage Examples

### Basic System Status
```yaml
tool: system-monitor
action: get_system_status
metrics:
  - "cpu_usage"
  - "memory_usage"
  - "disk_usage"
  - "network_io"
  - "load_average"
format: "json"
include_historical: false
```

### Process Monitoring
```yaml
tool: system-monitor
action: monitor_processes
filters:
  name_contains: ["python", "node"]
  memory_mb_gt: 100
  cpu_percent_gt: 50
output_format: "table"
refresh_interval: "5s"
duration: "60s"
```

### Health Check
```yaml
tool: system-monitor
action: perform_health_check
checks:
  - "disk_space"
  - "memory_pressure"
  - "cpu_thermal"
  - "network_connectivity"
  - "system_logs"
severity_threshold: "warning"
report_format: "detailed"
```

### Custom Alert Setup
```yaml
tool: system-monitor
action: setup_alerts
alerts:
  - metric: "memory_usage"
    threshold: 90
    condition: "greater_than"
    notification: "email"
    cooldown: "300s"
  - metric: "disk_usage"
    threshold: 85
    condition: "greater_than"
    notification: "internal"
    cooldown: "600s"
```

## Security Considerations

- System monitoring runs with minimal required privileges to prevent privilege escalation
- Sensitive system information is filtered based on access control policies
- Audit logging tracks all monitoring activities for security compliance
- Network monitoring respects privacy and data protection regulations
- Resource usage monitoring prevents excessive system load from monitoring itself

## Configuration

The system-monitor skill can be configured with the following parameters:

- `default_refresh_interval`: Default refresh interval for monitoring (default: 10s)
- `max_history_duration`: Maximum duration for historical data retention (default: 24h)
- `alert_thresholds`: Default alert thresholds for common metrics
- `privacy_level`: Privacy level for system information exposure (strict, moderate, relaxed)
- `monitoring_scope`: Scope of monitoring (basic, standard, comprehensive)

This skill is essential for any agent that needs to monitor system health, diagnose performance issues, track resource utilization, or ensure system reliability and availability.