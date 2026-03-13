---
name: docker-management
description: Comprehensive Docker container management system for AI agents with orchestration, monitoring, and security features
---

# Docker Management

This built-in skill provides comprehensive Docker container management capabilities for AI agents to deploy, manage, monitor, and secure containerized applications and services.

## Capabilities

- **Container Deployment**: Deploy and manage Docker containers from images or Dockerfiles
- **Container Orchestration**: Orchestrate multiple containers with networking, volumes, and dependencies
- **Image Management**: Build, pull, push, and manage Docker images with version control
- **Resource Monitoring**: Monitor CPU, memory, disk, and network usage of containers
- **Log Management**: Collect, analyze, and manage container logs with real-time streaming
- **Security Scanning**: Scan containers and images for vulnerabilities and security issues
- **Network Configuration**: Configure container networking, ports, and DNS settings
- **Volume Management**: Manage persistent volumes and data sharing between containers
- **Health Checks**: Implement and monitor container health checks and auto-recovery
- **Scaling and Load Balancing**: Scale containers horizontally and implement load balancing

## Usage Examples

### Deploy Container
```yaml
tool: docker-management
action: deploy_container
container:
  name: "web-app"
  image: "nginx:latest"
  ports:
    - "8080:80"
  volumes:
    - "/host/data:/container/data"
  environment:
    - "ENV=production"
    - "DEBUG=false"
  restart_policy: "unless-stopped"
  resource_limits:
    cpu: "2"
    memory: "2GB"
```

### Build Image
```yaml
tool: docker-management
action: build_image
image:
  name: "custom-app"
  tag: "v1.2.3"
  context: "/path/to/app"
  dockerfile: "Dockerfile"
  build_args:
    VERSION: "1.2.3"
    BUILD_ENV: "production"
  cache_from:
    - "custom-app:latest"
```

### Monitor Containers
```yaml
tool: docker-management
action: monitor_containers
filters:
  name: ["web-app", "database", "cache"]
metrics:
  - "cpu_usage"
  - "memory_usage"
  - "network_io"
  - "disk_io"
  - "log_volume"
alert_thresholds:
  cpu: 80
  memory: 90
  log_rate: "1000 lines/minute"
```

## Security Considerations

- Container images are scanned for vulnerabilities before deployment
- Network isolation prevents unauthorized container communication
- Resource limits prevent container resource exhaustion attacks
- Secure credential management handles Docker registry authentication
- Audit logging tracks all container management operations for compliance
- Principle of least privilege applied to container capabilities and permissions

## Configuration

The docker-management skill can be configured with the following parameters:

- `default_registry`: Default Docker registry (default: docker.io)
- `auto_scan_enabled`: Enable automatic vulnerability scanning (default: true)
- `log_retention`: Log retention period (default: 7 days)
- `health_check_interval`: Health check interval (default: 30s)
- `resource_limits_enabled`: Enable resource limits by default (default: true)
- `network_isolation_level`: Network isolation level (bridge, host, none, custom)

This skill is essential for any agent that needs to manage containerized applications, deploy microservices, or operate in cloud-native environments. It provides comprehensive Docker management capabilities while maintaining security and reliability.