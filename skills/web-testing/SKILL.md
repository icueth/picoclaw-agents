---
name: web-testing
description: Comprehensive web application testing and quality assurance system for AI agents with multi-framework support and automated test generation capabilities
---

# Web Testing

This built-in skill provides comprehensive web application testing and quality assurance capabilities for AI agents to create, execute, and manage automated tests for web applications across multiple frameworks and browsers.

## Capabilities

- **Multi-Framework Support**: Support for Jest, Cypress, Playwright, Selenium, Puppeteer, TestCafe, WebdriverIO, and Vitest
- **Test Type Coverage**: Support for unit, integration, end-to-end, visual regression, accessibility, and performance testing
- **Cross-Browser Testing**: Execute tests across multiple browsers (Chrome, Firefox, Safari, Edge) and devices
- **Intelligent Test Generation**: Automatically generate test cases from user stories, requirements, and application behavior
- **Visual Regression Testing**: Detect visual changes and regressions with pixel-perfect comparison
- **Accessibility Testing**: Ensure web applications meet accessibility standards (WCAG 2.1, Section 508)
- **Performance Testing**: Measure and analyze web application performance metrics (LCP, FID, CLS, TTFB)
- **Test Data Management**: Generate and manage realistic test data with edge cases and boundary conditions
- **CI/CD Integration**: Integrate with popular CI/CD platforms for automated test execution
- **Flaky Test Detection**: Identify and handle flaky tests with retry mechanisms and root cause analysis

## Usage Examples

### End-to-End Test Generation
```yaml
tool: web-testing
action: generate_e2e_tests
framework: "cypress"
application_url: "https://my-app.com"
test_scenarios:
  - name: "User Login Flow"
    steps:
      - "Navigate to login page"
      - "Enter valid credentials"
      - "Click login button"
      - "Verify dashboard is displayed"
  - name: "Product Search"
    steps:
      - "Navigate to homepage"
      - "Enter search term"
      - "Click search button"
      - "Verify search results are displayed"
output_directory: "/tests/e2e"
```

### Cross-Browser Testing
```yaml
tool: web-testing
action: execute_cross_browser_tests
framework: "playwright"
test_suite: "/tests/e2e"
browsers:
  - name: "chrome"
    version: "latest"
  - name: "firefox"
    version: "latest"
  - name: "safari"
    version: "latest"
  - name: "edge"
    version: "latest"
parallel_execution: true
max_retries: 2
report_format: "html"
```

### Visual Regression Testing
```yaml
tool: web-testing
action: perform_visual_regression
framework: "cypress"
test_pages:
  - url: "https://my-app.com/home"
    name: "homepage"
  - url: "https://my-app.com/dashboard"
    name: "dashboard"
  - url: "https://my-app.com/profile"
    name: "profile"
baseline_images: "/tests/visual/baseline"
threshold: 0.01
report_format: "html"
```

### Accessibility Testing
```yaml
<tool: web-testing
action: perform_accessibility_test
framework: "axe"
urls:
  - "https://my-app.com"
  - "https://my-app.com/about"
  - "https://my-app.com/contact"
standards:
  - "wcag2aa"
  - "section508"
severity_threshold: "serious"
report_format: "json"
include_fixes: true
```

## Security Considerations

- Test execution runs in isolated environments to prevent security vulnerabilities
- Test data is sanitized to remove sensitive information before use
- Access control ensures only authorized agents can execute tests or access test results
- Audit logging tracks all testing activities for compliance and security monitoring
- Security testing is integrated to detect vulnerabilities in the web application under test

## Configuration

The web-testing skill can be configured with the following parameters:

- `default_framework`: Default testing framework (cypress, playwright, jest, selenium)
- `max_parallel_executions`: Maximum number of parallel test executions (default: 10)
- `flaky_test_retry_count`: Number of retries for flaky tests (default: 3)
- `visual_regression_threshold`: Visual regression comparison threshold (default: 0.01)
- `accessibility_standards`: Default accessibility standards for testing (wcag2aa, section508)

This skill is essential for any agent that needs to ensure web application quality, automate testing workflows, validate user experiences, or integrate testing into development pipelines.