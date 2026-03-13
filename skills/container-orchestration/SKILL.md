---
name: container-orchestration
description: Advanced container orchestration and management system for AI agents with multi-platform support and intelligent scaling capabilities
---

# Container Orchestration

This built-in skill provides advanced container orchestration and management capabilities for AI agents to deploy, manage, and scale containerized applications across multiple platforms and environments.

## Capabilities

- **Multi-Platform Support**: Support for Kubernetes, Docker Swarm, OpenShift, Amazon ECS, and Google Cloud Run
- **Cluster Management**: Create, configure, and manage container orchestration clusters with automated provisioning
- **Application Deployment**: Deploy containerized applications with declarative configuration and rolling updates
- **Auto-Scaling**: Implement intelligent auto-scaling based on CPU, memory, custom metrics, and predictive analytics
- **Service Mesh Integration**: Integrate with service meshes (Istio, Linkerd, Consul) for advanced traffic management and observability
- **Security and Compliance**: Enforce security policies, network policies, and compliance requirements for containerized workloads
- **Resource Optimization**: Optimize resource allocation, scheduling, and cost management for container clusters
- **Monitoring and Logging**: Integrate with monitoring and logging solutions for comprehensive observability
- **Disaster Recovery**: Implement backup, restore, and disaster recovery strategies for containerized applications
- **GitOps Integration**: Support GitOps workflows with automated synchronization and drift detection

## Usage Examples

### Kubernetes Application Deployment
```yaml
tool: container-orchestration
action: deploy_application
platform: "kubernetes"
cluster: "production-cluster"
namespace: "my-app"
manifests:
  - "/k8s/deployment.yaml"
  - "/k8s/service.yaml"
  - "/k8s/ingress.yaml"
  - "/k8s/configmap.yaml"
strategy: "rolling_update"
health_checks:
  readiness_probe:
    http_get:
      path: "/health"
      port: 8080
    initial_delay_seconds: 10
    period_seconds: 5
  liveness_probe:
    http_get:
      path: "/live"
      port: 8080
    initial_delay_seconds: 30
    period_seconds: 10
```

### Intelligent Auto-Scaling
```yaml
tool: container-orchestration
action: configure_auto_scaling
platform: "kubernetes"
deployment: "web-app"
namespace: "production"
scaling_policies:
  - type: "horizontal_pod_autoscaler"
    metrics:
      - type: "cpu"
        target_average_utilization: 70
      - type: "memory"
        target_average_utilization: 80
      - type: "custom"
        name: "http_requests_per_second"
        target_value: "100"
    min_replicas: 2
    max_replicas: 10
  - type: "cluster_autoscaler"
    node_groups:
      - name: "spot-instances"
        min_size: 2
        max_size: 20
      - name: "on-demand-instances"
        min_size: 1
        max_size: 5
```

### Service Mesh Integration
```yaml
tool: container-orchestration
action: integrate_service_mesh
platform: "kubernetes"
service_mesh: "istio"
applications:
  - name: "frontend"
    namespace: "my-app"
    traffic_rules:
      - hosts: ["frontend.my-app.com"]
        gateways: ["my-app-gateway"]
        routes:
          - destination:
              host: "frontend"
              subset: "v1"
            weight: 90
          - destination:
              host: "frontend"
              subset: "v2"
            weight: 10
  - name: "backend"
    namespace: "my-app"
    security_policies:
      - peers:
          - mtls: {}
        port_level_mtls:
          "8080": "STRICT"
```

### GitOps Configuration
```yaml
tool: container-orchestration
action: configure_gitops
platform: "kubernetes"
gitops_tool: "argocd"
repository: "https://github.com/my-org/infra-config"
paths:
  - "clusters/production"
  - "applications/my-app"
sync_policy:
  automated: true
  prune: true
  self_heal: true
health_check:
  timeout: "300s"
  retry_limit: 5
notification_channels:
  - "slack"
  - "email"
```

## Security Considerations

- Container images are scanned for vulnerabilities before deployment
- Network policies restrict communication between pods and services
- Role-based access control (RBAC) enforces least privilege principles
- Secrets are encrypted and managed through secure secret management systems
- Audit logging tracks all orchestration activities for compliance and security monitoring

## Configuration

The container-orchestration skill can be configured with the following parameters:

- `default_platform`: Default orchestration platform (kubernetes, docker_swarm, ecs, cloud_run)
- `auto_scaling_enabled`: Enable auto-scaling by default (default: true)
- `security_policies`: Default security policies for containerized workloads
- `monitoring_integration`: Enabled monitoring and logging integrations
- `gitops_enabled`: Enable GitOps workflows by default (default: false)

This skill is essential for any agent that needs to manage containerized applications, implement scalable architectures, ensure security and compliance, or integrate modern DevOps practices into application deployment workflows.