---
name: system-hardening
description: Comprehensive system security hardening and vulnerability assessment for AI agents with automated security configuration and compliance checking
---

# System Hardening

This built-in skill provides comprehensive system security hardening and vulnerability assessment capabilities for AI agents to secure systems, detect vulnerabilities, and ensure compliance with security best practices.

## Capabilities

- **Security Configuration**: Apply security-hardened configurations to operating systems and applications
- **Vulnerability Scanning**: Scan systems for known vulnerabilities, misconfigurations, and security weaknesses
- **Compliance Checking**: Verify compliance with security standards (CIS, NIST, PCI-DSS, HIPAA)
- **Automated Hardening**: Automatically apply security patches and configuration changes
- **Access Control Management**: Configure and manage user accounts, permissions, and authentication
- **Network Security**: Harden network configurations, firewall rules, and service exposure
- **File System Security**: Secure file permissions, ownership, and access controls
- **Logging and Monitoring**: Configure secure logging and monitoring for security events
- **Threat Detection**: Detect and respond to potential security threats and intrusions
- **Security Reporting**: Generate comprehensive security reports and recommendations

## Usage Examples

### Basic System Hardening
```yaml
tool: system-hardening
action: harden_system
target: "current_host"
hardening_level: "standard"
exclude_services:
  - "ssh"
  - "web_server"
backup_config: true
revert_on_failure: true
```

### Vulnerability Assessment
```yaml
tool: system-hardening
action: assess_vulnerabilities
target: "192.168.1.0/24"
scan_types:
  - "os_vulnerabilities"
  - "misconfigurations"
  - "weak_passwords"
  - "open_ports"
severity_threshold: "medium"
include_fixes: true
```

### Compliance Check
```yaml
tool: system-hardening
action: check_compliance
target: "database_server"
standards:
  - "cis_mysql"
  - "pci_dss"
  - "hipaa"
report_format: "detailed"
include_remediation: true
```

### Automated Security Patching
```yaml
tool: system-hardening
action: apply_security_patches
target: "web_servers"
patch_level: "critical_and_high"
test_mode: false
backup_before_patch: true
rollback_on_failure: true
notification_recipients: ["security-team@example.com"]
```

## Security Considerations

- All hardening operations are performed with minimal required privileges
- Configuration changes are backed up before application to enable rollback
- Vulnerability scanning respects system resources and avoids denial-of-service
- Compliance checking uses official benchmark definitions and standards
- Security reporting redacts sensitive information while maintaining actionable insights

## Configuration

The system-hardening skill can be configured with the following parameters:

- `default_hardening_level`: Default hardening level (minimal, standard, maximum)
- `compliance_standards`: Enabled compliance standards (cis, nist, pci-dss, hipaa, soc2)
- `vulnerability_databases`: Vulnerability database sources (nvd, mitre, custom)
- `backup_retention`: Backup retention period for configuration changes (default: 30 days)
- `audit_logging_level`: Audit logging level for security operations (minimal, standard, comprehensive)

This skill is essential for any agent that needs to secure systems, maintain compliance with security standards, detect and remediate vulnerabilities, or implement enterprise-grade security hardening practices.