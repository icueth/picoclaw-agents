---
name: cloud-security
description: Comprehensive cloud security and compliance management system for AI agents with multi-platform support and automated threat detection capabilities
---

# Cloud Security

This built-in skill provides comprehensive cloud security and compliance management capabilities for AI agents to secure cloud infrastructure, detect threats, and ensure compliance with regulatory requirements across multiple platforms.

## Capabilities

- **Multi-Platform Support**: Secure AWS, Azure, Google Cloud, Kubernetes, and hybrid cloud environments
- **Infrastructure Security**: Scan and secure cloud infrastructure configurations for misconfigurations and vulnerabilities
- **Threat Detection**: Detect and respond to security threats, anomalies, and malicious activities in real-time
- **Compliance Management**: Ensure compliance with regulatory requirements (GDPR, HIPAA, PCI-DSS, SOC2, ISO 27001)
- **Identity and Access Management**: Manage and audit user permissions, roles, and access controls across cloud platforms
- **Data Protection**: Implement data encryption, tokenization, and privacy controls for sensitive information
- **Network Security**: Secure cloud networks with firewalls, security groups, and network segmentation
- **Security Automation**: Automate security responses, remediation, and incident handling workflows
- **Vulnerability Management**: Scan and manage vulnerabilities in cloud workloads, containers, and serverless functions
- **Security Monitoring**: Monitor security events and logs with intelligent alerting and correlation

## Usage Examples

### Cloud Infrastructure Security Scan
```yaml
tool: cloud-security
action: scan_infrastructure
platforms:
  - "aws"
  - "azure"
  - "gcp"
scan_types:
  - "misconfigurations"
  - "security_vulnerabilities"
  - "compliance_violations"
  - "excessive_permissions"
compliance_standards:
  - "cis_aws"
  - "cis_azure"
  - "cis_gcp"
  - "pci_dss"
  - "hipaa"
severity_threshold: "medium"
auto_remediate: true
```

### Threat Detection and Response
```yaml
tool: cloud-security
action: detect_threats
platforms:
  - "aws"
  - "kubernetes"
threat_types:
  - "unauthorized_access"
  - "data_exfiltration"
  - "malware_activity"
  - "privilege_escalation"
  - "suspicious_network_traffic"
detection_methods:
  - "anomaly_detection"
  - "behavioral_analysis"
  - "signature_based"
  - "machine_learning"
response_actions:
  - type: "isolate_resource"
    condition: "threat_severity = 'critical'"
  - type: "notify_security_team"
    condition: "threat_severity >= 'high'"
  - type: "block_ip_address"
    condition: "threat_type = 'unauthorized_access'"
```

### Identity and Access Management Audit
```yaml
tool: cloud-security
action: audit_iam
platforms:
  - "aws"
  - "azure"
  - "gcp"
audit_types:
  - "excessive_permissions"
  - "unused_credentials"
  - "privileged_accounts"
  - "mfa_compliance"
  - "role_assignments"
remediation_rules:
  - condition: "unused_for > 90_days"
    action: "disable_credential"
  - condition: "admin_privileges = true AND mfa_enabled = false"
    action: "require_mfa"
  - condition: "permission_scope = 'wildcard'"
    action: "restrict_permissions"
```

### Compliance Management
```yaml
tool: cloud-security
action: manage_compliance
compliance_frameworks:
  - "gdpr"
  - "hipaa"
  - "pci_dss"
  - "soc2"
  - "iso_27001"
assessment_scope:
  - "aws_accounts"
  - "azure_subscriptions"
  - "gcp_projects"
  - "kubernetes_clusters"
controls:
  - "data_encryption"
  - "access_controls"
  - "audit_logging"
  - "incident_response"
  - "vulnerability_management"
reporting_frequency: "monthly"
notification_channels: ["email", "slack"]
```

## Security Considerations

- Security scans and assessments run with minimal required permissions to prevent privilege escalation
- Sensitive security findings are encrypted and access-controlled
- Automated remediation actions are validated and require appropriate approvals
- Audit logging tracks all security activities for compliance and forensic investigations
- Threat intelligence feeds are regularly updated from trusted sources

## Configuration

The cloud-security skill can be configured with the following parameters:

- `default_platforms`: Default cloud platforms to monitor (aws, azure, gcp, kubernetes)
- `compliance_frameworks`: Enabled compliance frameworks for assessment
- `auto_remediation_level`: Level of automated remediation (none, low_risk, medium_risk, all)
- `threat_intelligence_sources`: Enabled threat intelligence sources
- `audit_logging_retention`: Audit log retention period (default: 365 days)

This skill is essential for any agent that needs to secure cloud infrastructure, detect and respond to threats, ensure regulatory compliance, or implement enterprise-grade security practices in cloud environments.