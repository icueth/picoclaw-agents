---
name: input-validator
description: Comprehensive input validation and sanitization system for AI agents with security-focused validation rules and threat detection
---

# Input Validator

This built-in skill provides comprehensive input validation and sanitization capabilities for AI agents to ensure data integrity, prevent security vulnerabilities, and maintain system reliability.

## Capabilities

- **Data Type Validation**: Validate inputs against expected data types (string, number, boolean, array, object)
- **Format Validation**: Validate inputs against specific formats (email, URL, phone, date, regex patterns)
- **Length and Range Validation**: Enforce length limits, numeric ranges, and size constraints
- **Security Validation**: Detect and prevent common security threats (SQL injection, XSS, command injection)
- **Custom Validation Rules**: Define custom validation logic with complex business rules
- **Input Sanitization**: Clean and sanitize inputs to remove malicious content while preserving valid data
- **Batch Validation**: Validate multiple inputs simultaneously with consistent error reporting
- **Schema Validation**: Validate inputs against JSON Schema, OpenAPI, or custom schema definitions
- **Threat Intelligence**: Integrate with threat intelligence feeds for real-time threat detection
- **Audit Logging**: Log all validation activities for security compliance and debugging

## Usage Examples

### Basic Input Validation
```yaml
tool: input-validator
action: validate_input
input: "{{user_email}}"
validation_rules:
  - type: "string"
  - format: "email"
  - max_length: 254
  - required: true
error_handling: "reject"
```

### Security-Focused Validation
```yaml
tool: input-validator
action: validate_secure_input
input: "{{user_query}"
validation_rules:
  - type: "string"
  - max_length: 1000
  - no_sql_injection: true
  - no_xss: true
  - no_command_injection: true
  - no_path_traversal: true
  - allowed_characters: "alphanumeric_plus_safe_symbols"
sanitization: "aggressive"
```

### Schema Validation
```yaml
tool: input-validator
action: validate_schema
input: "{{user_data}"
schema:
  type: "object"
  properties:
    name:
      type: "string"
      minLength: 1
      maxLength: 100
    age:
      type: "integer"
      minimum: 0
      maximum: 150
    email:
      type: "string"
      format: "email"
    preferences:
      type: "array"
      items:
        type: "string"
        enum: ["email", "sms", "push"]
  required: ["name", "email"]
strict: true
```

### Batch Validation
```yaml
tool: input-validator
action: batch_validate
inputs:
  username: "{{username}}"
  password: "{{password}}"
  email: "{{email}}"
  phone: "{{phone}}"
validation_rules:
  username:
    - type: "string"
    - min_length: 3
    - max_length: 20
    - pattern: "^[a-zA-Z0-9_]+$"
  password:
    - type: "string"
    - min_length: 8
    - complexity: "high"
  email:
    - type: "string"
    - format: "email"
  phone:
    - type: "string"
    - format: "phone"
    - country_code: "US"
error_handling: "collect_all"
```

## Security Considerations

- All validation rules are executed in isolated environments to prevent bypass attempts
- Input sanitization uses proven libraries and techniques to prevent security vulnerabilities
- Threat detection integrates with up-to-date security databases and patterns
- Audit logging provides complete traceability for security investigations
- Validation failures are handled securely to prevent information disclosure

## Configuration

The input-validator skill can be configured with the following parameters:

- `default_validation_level`: Default validation strictness (strict, moderate, relaxed)
- `security_rules_enabled`: Enable security-focused validation rules (default: true)
- `max_input_size`: Maximum input size for validation (default: 10KB)
- `threat_intelligence_sources`: Enabled threat intelligence sources
- `audit_logging_level`: Audit logging level (minimal, standard, comprehensive)

This skill is essential for any agent that needs to process user input, validate data from external sources, prevent security vulnerabilities, or ensure data integrity across automated workflows.