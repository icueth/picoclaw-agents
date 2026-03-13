---
name: agent-security-check
description: Comprehensive security scanning and validation for AI agents and their operations
---

# Agent Security Check

This built-in skill provides comprehensive security scanning and validation capabilities for AI agents to ensure safe operations, detect vulnerabilities, and prevent malicious activities.

## Capabilities

- **Code Scanning**: Scan code for security vulnerabilities, hardcoded secrets, and dangerous patterns
- **Input Validation**: Validate inputs for injection attacks, XSS, and other common vulnerabilities
- **Dependency Analysis**: Analyze dependencies for known vulnerabilities (CVE database integration)
- **Network Security**: Monitor network requests for suspicious patterns or data exfiltration
- **File System Security**: Scan file operations for unauthorized access or dangerous operations
- **Prompt Injection Detection**: Detect and prevent prompt injection attacks in agent interactions
- **Malware Scanning**: Scan downloaded files and content for malware signatures
- **Compliance Checking**: Verify compliance with security best practices and standards
- **Threat Intelligence**: Integrate with threat intelligence feeds for real-time protection
- **Security Auditing**: Generate comprehensive security audit reports and recommendations

## Usage Examples

### Code Security Scan
```yaml
tool: agent-security-check
action: scan_code
path: "/project/src"
scan_type: "comprehensive"
include_secrets: true
include_vulnerabilities: true
```

### Input Validation
```yaml
tool: agent-security-check
action: validate_input
input: "{{user_input}}"
validation_rules:
  - "no_sql_injection"
  - "no_xss"
  - "no_command_injection"
  - "length_limit:1000"
```

### Dependency Vulnerability Scan
```yaml
tool: agent-security-check
action: scan_dependencies
manifest_files:
  - "package.json"
  - "go.mod"
  - "requirements.txt"
severity_threshold: "high"
```

## Security Considerations

- All security checks run in isolated sandboxed environments
- No sensitive data is transmitted to external services without explicit consent
- Security rules are regularly updated from trusted sources
- False positive reduction mechanisms minimize disruption to legitimate operations
- Comprehensive logging for security incident investigation

## Configuration

The agent-security-check skill can be configured with the following parameters:

- `scan_depth`: Depth of security scanning (shallow, deep, comprehensive)
- `severity_threshold`: Minimum severity level to report (low, medium, high, critical)
- `auto_remediate`: Automatically fix low-risk issues (default: false)
- `external_integrations`: External security services to integrate with
- `scan_schedule`: Automated scanning schedule for continuous protection
- `exclusion_patterns`: File patterns to exclude from scanning

This skill is essential for any agent that handles code, processes user input, manages dependencies, or performs operations that could impact system security. It provides a critical safety net for AI agent operations.