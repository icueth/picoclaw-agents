---
name: code-analyzer
description: Advanced code analysis and static analysis system for AI agents with multi-language support and security-focused insights
---

# Code Analyzer

This built-in skill provides advanced code analysis and static analysis capabilities for AI agents to analyze source code, detect issues, and provide actionable insights across multiple programming languages.

## Capabilities

- **Multi-Language Support**: Analyze code in over 50 programming languages (Python, JavaScript, Java, Go, Rust, C++, etc.)
- **Static Analysis**: Perform comprehensive static analysis to detect bugs, performance issues, and code smells
- **Security Scanning**: Identify security vulnerabilities, hardcoded secrets, and dangerous patterns
- **Code Quality Metrics**: Calculate code quality metrics (cyclomatic complexity, maintainability index, code coverage)
- **Dependency Analysis**: Analyze dependencies for known vulnerabilities and license compliance
- **Architecture Analysis**: Analyze code architecture, dependencies, and module relationships
- **Style Compliance**: Check code against style guides and best practices (PEP 8, Google Style, Airbnb, etc.)
- **Technical Debt Assessment**: Identify and quantify technical debt with remediation recommendations
- **Performance Analysis**: Detect performance bottlenecks and optimization opportunities
- **Automated Refactoring**: Suggest and apply automated refactoring improvements

## Usage Examples

### Basic Code Analysis
```yaml
tool: code-analyzer
action: analyze_code
path: "/project/src"
languages: ["python", "javascript"]
analysis_types:
  - "static_analysis"
  - "security_scanning"
  - "code_quality"
  - "style_compliance"
severity_threshold: "medium"
output_format: "detailed"
```

### Security-Focused Analysis
```yaml
tool: code-analyzer
action: security_scan
path: "/project"
scan_types:
  - "hardcoded_secrets"
  - "sql_injection"
  - "xss_vulnerabilities"
  - "command_injection"
  - "insecure_dependencies"
  - "crypto_weaknesses"
exclude_patterns:
  - "node_modules/"
  - "venv/"
  - "*.min.js"
report_format: "sarif"
```

### Dependency Vulnerability Scan
```yaml
tool: code-analyzer
action: scan_dependencies
manifest_files:
  - "package.json"
  - "requirements.txt"
  - "go.mod"
  - "pom.xml"
vulnerability_databases:
  - "nvd"
  - "github_advisories"
  - "snyk"
severity_threshold: "high"
include_fixes: true
```

### Architecture Analysis
```yaml
tool: code-analyzer
action: analyze_architecture
path: "/project"
analysis_types:
  - "dependency_graph"
  - "circular_dependencies"
  - "module_coupling"
  - "layer_violations"
output_formats:
  - "graphviz"
  - "json"
  - "interactive_html"
```

## Security Considerations

- Code analysis runs in isolated sandboxed environments to prevent code execution
- Sensitive code repositories are never transmitted to external services
- Access control ensures only authorized agents can analyze specific codebases
- Audit logging tracks all code analysis activities for compliance and security
- Vulnerability databases are regularly updated from trusted sources

## Configuration

The code-analyzer skill can be configured with the following parameters:

- `default_languages`: Default programming languages to analyze
- `analysis_depth`: Analysis depth (shallow, deep, comprehensive)
- `security_rules_enabled`: Enable security-focused analysis rules (default: true)
- `vulnerability_databases`: Enabled vulnerability database sources
- `output_formats`: Supported output formats (json, sarif, html, markdown)

This skill is essential for any agent that needs to analyze code quality, detect security vulnerabilities, assess technical debt, or provide actionable insights for code improvement and maintenance.