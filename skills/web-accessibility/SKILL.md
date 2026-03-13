---
name: web-accessibility
description: Comprehensive web accessibility testing and compliance system for AI agents with multi-standard support and automated remediation capabilities
---

# Web Accessibility

This built-in skill provides comprehensive web accessibility testing and compliance capabilities for AI agents to ensure web applications meet accessibility standards and provide inclusive user experiences.

## Capabilities

- **Multi-Standard Support**: Support for WCAG 2.1/2.2, Section 508, ADA, EN 301 549, and ARIA guidelines
- **Automated Testing**: Perform automated accessibility testing with comprehensive rule coverage and false positive reduction
- **Manual Testing Guidance**: Provide guidance for manual accessibility testing and user testing with assistive technologies
- **Compliance Reporting**: Generate detailed compliance reports with violation severity, impact, and remediation guidance
- **Automated Remediation**: Apply automated fixes for common accessibility issues with safe rollback mechanisms
- **Screen Reader Testing**: Simulate screen reader experiences and test keyboard navigation
- **Color Contrast Analysis**: Analyze color contrast ratios and provide accessible color palette recommendations
- **Semantic HTML Validation**: Validate semantic HTML structure and proper ARIA attribute usage
- **Focus Management**: Test and validate keyboard focus management and logical tab order
- **Continuous Monitoring**: Monitor accessibility compliance continuously with alerts and trend analysis

## Usage Examples

### Comprehensive Accessibility Audit
```yaml
tool: web-accessibility
action: audit_accessibility
url: "https://my-app.com"
standards:
  - "wcag2aa"
  - "section508"
  - "ada"
device_types:
  - "desktop"
  - "mobile"
  - "tablet"
assistive_technologies:
  - "screen_reader"
  - "keyboard_only"
  - "voice_navigation"
severity_threshold: "serious"
include_remediation: true
report_format: "detailed"
```

### Automated Remediation
```yaml
tool: web-accessibility
action: remediate_issues
url: "https://my-app.com"
issues_to_fix:
  - "missing_alt_text"
  - "insufficient_color_contrast"
  - "missing_form_labels"
  - "empty_links"
  - "missing_aria_roles"
auto_approve: false
preview_changes: true
backup_original: true
```

### Color Contrast Analysis
```yaml
tool: web-accessibility
action: analyze_color_contrast
url: "https://my-app.com"
contrast_requirements:
  - "aa_normal_text"
  - "aa_large_text"
  - "aaa_normal_text"
  - "aaa_large_text"
generate_accessible_palette: true
output_format: "css"
```

### Keyboard Navigation Testing
```yaml
tool: web-accessibility
action: test_keyboard_navigation
url: "https://my-app.com"
test_scenarios:
  - name: "Main Navigation"
    expected_path: ["skip_link", "nav_menu", "main_content", "footer"]
  - name: "Form Completion"
    expected_path: ["first_name", "last_name", "email", "submit_button"]
  - name: "Modal Dialog"
    expected_path: ["open_modal", "modal_content", "close_modal"]
focus_indicators: true
trap_testing: true
```

## Security Considerations

- Accessibility testing runs in isolated environments to prevent data leakage
- Sensitive accessibility findings are encrypted and access-controlled
- Automated remediation changes are validated with safety checks before application
- Access control ensures only authorized agents can modify accessibility configurations
- Audit logging tracks all accessibility testing and remediation activities for compliance

## Configuration

The web-accessibility skill can be configured with the following parameters:

- `default_standards`: Default accessibility standards (wcag2aa, section508, ada)
- `severity_threshold`: Minimum severity level for reporting (minor, moderate, serious, critical)
- `auto_remediation_level`: Level of automated remediation (none, low_risk, medium_risk, all)
- `monitoring_frequency`: Frequency of accessibility monitoring (real_time, daily, weekly)
- `compliance_reporting`: Enable compliance reporting for regulatory requirements (default: true)

This skill is essential for any agent that needs to ensure web application accessibility, comply with legal requirements, provide inclusive user experiences, or implement accessibility best practices in development workflows.