---
name: agent-rate-limiter
description: Intelligent rate limiting and throttling system for AI agents to prevent API abuse and ensure fair resource usage
---

# Agent Rate Limiter

This built-in skill provides intelligent rate limiting and throttling capabilities for AI agents to prevent API abuse, ensure fair resource usage, and maintain system stability.

## Capabilities

- **API Rate Limiting**: Enforce rate limits on external API calls with configurable thresholds
- **Dynamic Throttling**: Automatically adjust request rates based on response patterns and errors
- **Quota Management**: Track and manage usage quotas across different services and time periods
- **Exponential Backoff**: Implement exponential backoff strategies for failed requests
- **Concurrent Request Control**: Limit concurrent requests to prevent overwhelming services
- **Token Bucket Algorithm**: Use token bucket algorithm for smooth rate limiting
- **Service-Specific Rules**: Apply different rate limiting rules based on service characteristics
- **Real-time Monitoring**: Monitor rate limit usage and provide alerts for approaching limits
- **Graceful Degradation**: Implement graceful degradation when limits are reached
- **Historical Analysis**: Analyze historical usage patterns to optimize rate limiting policies

## Usage Examples

### Basic Rate Limiting
```yaml
tool: agent-rate-limiter
action: enforce_limit
service: "github_api"
max_requests: 5000
time_window: "3600s"
strategy: "token_bucket"
```

### Dynamic Throttling
```yaml
tool: agent-rate-limiter
action: apply_throttling
service: "openai_api"
base_rate: 10
error_multiplier: 2.0
success_multiplier: 0.9
min_rate: 1
max_rate: 50
```

### Quota Management
```yaml
tool: agent-rate-limiter
action: check_quota
service: "custom_service"
quota_type: "daily"
quota_limit: 10000
current_usage: "{{current_usage}}"
```

## Security Considerations

- Rate limiting prevents denial-of-service attacks through API abuse
- Quota management ensures fair resource distribution among users
- Real-time monitoring detects unusual usage patterns that may indicate compromise
- Graceful degradation maintains service availability during high load
- Audit logging tracks all rate limiting decisions for compliance

## Configuration

The agent-rate-limiter skill can be configured with the following parameters:

- `default_strategy`: Default rate limiting strategy (token_bucket, leaky_bucket, fixed_window)
- `global_max_concurrent`: Global maximum concurrent requests (default: 100)
- `service_configs`: Service-specific configuration overrides
- `alert_threshold`: Percentage of quota usage that triggers alerts (default: 80%)
- `backoff_strategy`: Backoff strategy for failed requests (exponential, linear, random)
- `monitoring_enabled`: Enable real-time monitoring and alerting (default: true)

This skill is essential for any agent that makes external API calls, manages multiple services, or needs to ensure stable and fair resource usage. It protects both the agent and the services it interacts with from abuse and overload.