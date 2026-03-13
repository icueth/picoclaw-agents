---
name: testing-automation
description: Comprehensive test automation and quality assurance system for AI agents with multi-framework support and intelligent test generation capabilities
---

# Testing Automation

This built-in skill provides comprehensive test automation and quality assurance capabilities for AI agents to create, execute, and manage automated tests across multiple frameworks and testing types.

## Capabilities

- **Multi-Framework Support**: Support for popular testing frameworks (JUnit, pytest, Jest, Mocha, Cypress, Selenium, Playwright, etc.)
- **Test Type Coverage**: Support for unit, integration, end-to-end, API, performance, and security testing
- **Intelligent Test Generation**: Automatically generate test cases from code, requirements, and user stories
- **Test Data Management**: Generate and manage test data with realistic scenarios and edge cases
- **Test Execution Orchestration**: Orchestrate test execution across multiple environments and browsers
- **Flaky Test Detection**: Identify and handle flaky tests with retry mechanisms and root cause analysis
- **Code Coverage Analysis**: Measure and analyze code coverage with detailed reports and recommendations
- **Test Result Analysis**: Analyze test results to identify patterns, trends, and quality issues
- **CI/CD Integration**: Integrate with popular CI/CD platforms (GitHub Actions, Jenkins, GitLab CI, CircleCI)
- **Quality Gates**: Implement quality gates and thresholds for automated decision making

## Usage Examples

### Intelligent Test Generation
```yaml
tool: testing-automation
action: generate_tests
source:
  type: "code"
  path: "/src/calculator.js"
test_framework: "jest"
test_types:
  - "unit"
  - "integration"
coverage_target: 90
include_edge_cases: true
output_directory: "/tests/generated"
```

### Multi-Environment Test Execution
```yaml
tool: testing-automation
action: execute_tests
test_suite: "/tests/e2e"
environments:
  - name: "chrome_latest"
    browser: "chrome"
    version: "latest"
  - name: "firefox_latest"
    browser: "firefox"
    version: "latest"
  - name: "safari_latest"
    browser: "safari"
    version: "latest"
parallel_execution: true
max_retries: 2
report_format: "html"
```

### Code Coverage Analysis
```yaml
tool: testing-automation
action: analyze_coverage
source_code: "/src"
test_results: "/reports/coverage.xml"
coverage_thresholds:
  statements: 80
  branches: 75
  functions: 85
  lines: 80
report_formats:
  - "html"
  - "json"
  - "lcov"
recommendations: true
```

### Quality Gate Implementation
```yaml
tool: testing-automation
action: implement_quality_gate
gate_rules:
  - metric: "test_pass_rate"
    threshold: 95
    operator: "greater_than_or_equal"
  - metric: "code_coverage"
    threshold: 80
    operator: "greater_than_or_equal"
  - metric: "security_vulnerabilities"
    threshold: 0
    operator: "equal"
  - metric: "performance_regression"
    threshold: 10
    operator: "less_than_or_equal"
failure_action: "block_merge"
notification_channels: ["slack", "email"]
```

## Security Considerations

- Test execution runs in isolated environments to prevent security vulnerabilities
- Test data is sanitized to remove sensitive information before use
- Access control ensures only authorized agents can execute tests or access test results
- Audit logging tracks all test automation activities for compliance and security monitoring
- Security testing is integrated to detect vulnerabilities in the application under test

## Configuration

The testing-automation skill can be configured with the following parameters:

- `default_test_frameworks`: Default testing frameworks by language
- `max_parallel_executions`: Maximum number of parallel test executions (default: 10)
- `flaky_test_retry_count`: Number of retries for flaky tests (default: 3)
- `coverage_thresholds`: Default code coverage thresholds by metric type
- `ci_cd_integrations`: Enabled CI/CD platform integrations

This skill is essential for any agent that needs to automate software testing, ensure code quality, implement quality gates, or integrate testing into development workflows.