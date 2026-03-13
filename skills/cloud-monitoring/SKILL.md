---
name: cloud-monitoring
description: Comprehensive cloud monitoring and observability system for AI agents with multi-platform support and intelligent alerting capabilities
---

# Cloud Monitoring

This built-in skill provides comprehensive cloud monitoring and observability capabilities for AI agents to monitor, analyze, and optimize cloud infrastructure and applications across multiple platforms and services.

## Capabilities

- **Multi-Platform Support**: Monitor AWS, Azure, Google Cloud, Kubernetes, and on-premises infrastructure
- **Metrics Collection**: Collect and analyze metrics for CPU, memory, disk, network, and application performance
- **Log Aggregation**: Aggregate and analyze logs from multiple sources with intelligent parsing and filtering
- **Distributed Tracing**: Track requests across microservices with distributed tracing and performance analysis
- **Intelligent Alerting**: Create intelligent alerts with anomaly detection, machine learning, and dynamic thresholds
- **Dashboard Management**: Create and manage interactive dashboards for real-time monitoring and visualization
- **Incident Response**: Automate incident response workflows with escalation, notification, and remediation
- **Cost Monitoring**: Monitor and optimize cloud costs with budget alerts and optimization recommendations
- **Security Monitoring**: Detect security threats and anomalies with integrated security monitoring
- **Custom Metrics**: Define and track custom business and application metrics

## Usage Examples

### Multi-Cloud Metrics Monitoring
```yaml
tool: cloud-monitoring
action: monitor_metrics
platforms:
  - "aws"
  - "azure"
  - "gcp"
  - "kubernetes"
metrics:
  - "cpu_utilization"
  - "memory_usage"
  - "disk_io"
  - "network_traffic"
  - "application_latency"
  - "error_rate"
aggregation: "average"
time_range: "1h"
alert_thresholds:
  cpu_utilization: 80
  memory_usage: 85
  error_rate: 0.01
```

### Log Analysis and Alerting
```yaml
tool: cloud-monitoring
action: analyze_logs
log_sources:
  - "/var/log/application/*.log"
  - "cloudwatch:my-app-logs"
  - "kubernetes:namespace=my-app"
filters:
  - "level = 'ERROR'"
  - "service = 'payment'"
  - "timestamp > '2026-03-13T00:00:00Z'"
analysis_types:
  - "error_patterns"
  - "performance_bottlenecks"
  - "security_anomalies"
alert_on:
  - "error_count > 100/hour"
  - "latency_p95 > 2000ms"
```

### Distributed Tracing
```yaml
tool: cloud-monitoring
action: trace_requests
service_name: "order-processing"
trace_types:
  - "http_requests"
  - "database_queries"
  - "external_api_calls"
  - "message_queue_operations"
time_range: "30m"
performance_thresholds:
  duration_ms: 1000
  database_queries: 10
  external_calls: 5
```

### Intelligent Incident Response
```yaml
tool: cloud-monitoring
action: create_incident_response
alert_name: "High CPU Utilization"
conditions:
  - "cpu_utilization > 90"
  - "duration > 5m"
actions:
  - type: "notification"
    channels: ["slack", "email", "pagerduty"]
    recipients: ["devops-team"]
  - type: "auto_remediation"
    action: "scale_instances"
    parameters:
      min_instances: 2
      max_instances: 10
  - type: "run_diagnostics"
    script: "/scripts/cpu_diagnostics.sh"
escalation:
  - delay: "15m"
    action: "notify_manager"
```

## Security Considerations

- Monitoring data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access sensitive monitoring data
- Alert notifications are secured to prevent unauthorized access to incident details
- Log data is filtered to remove sensitive information before analysis
- Audit logging tracks all monitoring activities for compliance and security monitoring

## Configuration

The cloud-monitoring skill can be configured with the following parameters:

- `default_platforms`: Default monitoring platforms (aws, azure, gcp, kubernetes)
- `data_retention_period`: Data retention period for metrics and logs (default: 30 days)
- `alert_severity_levels`: Alert severity levels and corresponding actions
- `notification_channels`: Enabled notification channels (slack, email, pagerduty, webhook)
- `privacy_filtering`: Privacy filtering rules for sensitive data in logs

This skill is essential for any agent that needs to monitor cloud infrastructure, detect and respond to incidents, optimize performance and costs, or ensure application reliability and availability.