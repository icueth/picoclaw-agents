---
name: web-security
description: Comprehensive web application security testing and protection system for AI agents with multi-layered security analysis and automated vulnerability remediation capabilities
---

# Web Security

This built-in skill provides comprehensive web application security testing and protection capabilities for AI agents to identify, analyze, and remediate security vulnerabilities across multiple layers of web applications.

## Capabilities

- **Vulnerability Scanning**: Scan for common web vulnerabilities (OWASP Top 10) including XSS, SQL injection, CSRF, SSRF, and more
- **Authentication Testing**: Test authentication mechanisms, session management, and password policies
- **Authorization Testing**: Verify proper access controls, role-based permissions, and privilege escalation prevention
- **Input Validation Testing**: Test input validation, sanitization, and output encoding across all application endpoints
- **Security Headers Analysis**: Analyze and enforce proper security headers (CSP, HSTS, X-Frame-Options, etc.)
- **API Security Testing**: Test REST and GraphQL APIs for security vulnerabilities and proper authentication
- **Client-Side Security**: Analyze client-side code for security issues (insecure JavaScript, DOM-based XSS, etc.)
- **Automated Remediation**: Apply automated fixes for common security vulnerabilities with safe rollback mechanisms
- **Compliance Checking**: Verify compliance with security standards (OWASP ASVS, PCI-DSS, SOC2)
- **Continuous Monitoring**: Monitor web applications continuously for security issues with real-time alerts

## Usage Examples

### Comprehensive Security Scan
```yaml
tool: web-security
action: scan_vulnerabilities
url: "https://my-app.com"
scan_types:
  - "xss"
  - "sql_injection"
  - "csrf"
  - "ssrf"
  - "command_injection"
  - "path_traversal"
  - "insecure_direct_object_references"
  - "security_misconfigurations"
  - "sensitive_data_exposure"
  - "broken_authentication"
depth: "comprehensive"
crawl_depth: 5
include_subdomains: true
severity_threshold: "medium"
```

### Authentication Security Testing
```yaml
tool: web-security
action: test_authentication
url: "https://my-app.com"
test_types:
  - "brute_force_protection"
  - "account_lockout"
  - "password_policy"
  - "session_timeout"
  - "session_fixation"
  - "jwt_security"
  - "oauth_implementation"
  - "multi_factor_authentication"
credentials:
  valid_user: "test@example.com"
  valid_password: "{{test_password}}"
  invalid_attempts: 10
```

### Security Headers Analysis
```yaml
tool: web-security
action: analyze_security_headers
url: "https://my-app.com"
required_headers:
  - "Content-Security-Policy"
  - "Strict-Transport-Security"
  - "X-Frame-Options"
  - "X-Content-Type-Options"
  - "Referrer-Policy"
  - "Permissions-Policy"
csp_analysis: true
hsts_analysis: true
recommendations: true
```

### API Security Testing
```yaml
tool: web-security
action: test_api_security
api_spec: "/openapi.json"
endpoints:
  - "/api/users"
  - "/api/orders"
  - "/api/payments"
test_types:
  - "authentication_bypass"
  - "authorization_bypass"
  - "input_validation"
  - "rate_limiting"
  - "data_exfiltration"
  - "business_logic_abuse"
auth_methods:
  - "bearer_token"
  - "api_key"
  - "oauth2"
```

## Security Considerations

- Security testing runs in isolated environments to prevent actual exploitation
- Sensitive security findings are encrypted and access-controlled
- Automated remediation changes are validated with safety checks before application
- Access control ensures only authorized agents can perform security testing
- Audit logging tracks all security testing and remediation activities for compliance
- False positive reduction mechanisms minimize disruption to legitimate operations

## Configuration

The web-security skill can be configured with the following parameters:

- `default_scan_depth`: Default scan depth (shallow, moderate, comprehensive)
- `severity_threshold`: Minimum severity level for reporting (low, medium, high, critical)
- `auto_remediation_level`: Level of automated remediation (none, low_risk, medium_risk, all)
- `compliance_standards`: Enabled compliance standards (owasp_asvs, pci_dss, soc2)
- `monitoring_frequency`: Frequency of security monitoring (real_time, daily, weekly)

This skill is essential for any agent that needs to secure web applications, identify and fix vulnerabilities, ensure compliance with security standards, or implement secure development practices in web development workflows.