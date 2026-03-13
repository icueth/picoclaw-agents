---
name: code-refactoring
description: Intelligent code refactoring and optimization system for AI agents with multi-language support and automated improvement capabilities
---

# Code Refactoring

This built-in skill provides intelligent code refactoring and optimization capabilities for AI agents to improve code quality, maintainability, and performance across multiple programming languages.

## Capabilities

- **Multi-Language Support**: Refactor code in over 50 programming languages (Python, JavaScript, Java, Go, Rust, C++, etc.)
- **Automated Refactoring**: Automatically apply safe refactoring transformations (rename, extract method, inline variable, etc.)
- **Code Quality Improvement**: Identify and fix code smells, anti-patterns, and maintainability issues
- **Performance Optimization**: Optimize code for better performance, memory usage, and resource efficiency
- **Architecture Refactoring**: Restructure code architecture, modules, and dependencies for better design
- **Technical Debt Reduction**: Identify and reduce technical debt with systematic refactoring approaches
- **Style Standardization**: Apply consistent coding styles and best practices across codebases
- **API Evolution**: Safely evolve APIs while maintaining backward compatibility
- **Dependency Management**: Refactor dependency structures and manage third-party library updates
- **Refactoring Validation**: Validate refactoring changes with automated tests and static analysis

## Usage Examples

### Automated Code Refactoring
```yaml
tool: code-refactoring
action: refactor_code
path: "/src/utils.js"
refactorings:
  - type: "extract_function"
    start_line: 25
    end_line: 35
    new_function_name: "validateUserInput"
  - type: "rename_variable"
    old_name: "tmp"
    new_name: "processedData"
  - type: "inline_variable"
    variable_name: "intermediateResult"
language: "javascript"
preview_changes: true
validate_with_tests: true
```

### Code Quality Improvement
```yaml
tool: code-refactoring
action: improve_quality
path: "/src"
improvements:
  - "eliminate_duplicate_code"
  - "reduce_cyclomatic_complexity"
  - "improve_naming_conventions"
  - "fix_code_smells"
  - "apply_solid_principles"
severity_threshold: "medium"
exclude_patterns:
  - "node_modules/"
  - "test/"
output_format: "detailed"
```

### Performance Optimization
```yaml
tool: code-refactoring
action: optimize_performance
path: "/src/data_processor.py"
optimizations:
  - "eliminate_n_plus_1_queries"
  - "cache_expensive_operations"
  - "optimize_data_structures"
  - "reduce_memory_allocations"
  - "parallelize_computation"
benchmark_before: true
benchmark_after: true
performance_threshold: 20
```

### Architecture Refactoring
```yaml
tool: code-refactoring
action: refactor_architecture
path: "/src"
refactorings:
  - type: "extract_module"
    source_files: ["/src/core.js", "/src/utils.js"]
    target_module: "core"
  - type: "eliminate_circular_dependency"
    module_a: "services"
    module_b: "models"
  - type: "apply_layered_architecture"
    layers: ["presentation", "business", "data"]
validation_strategy: "comprehensive"
```

## Security Considerations

- Refactoring operations are validated against security policies before execution
- Code changes are previewed and validated with automated tests before application
- Access control ensures only authorized agents can modify codebases
- Audit logging tracks all refactoring activities for compliance and security monitoring
- Sensitive code patterns are preserved during refactoring to maintain security

## Configuration

The code-refactoring skill can be configured with the following parameters:

- `default_languages`: Default programming languages for refactoring support
- `refactoring_safety_level`: Safety level for automated refactoring (conservative, moderate, aggressive)
- `max_file_size`: Maximum file size for refactoring operations (default: 5MB)
- `validation_strategies`: Enabled validation strategies (tests, static_analysis, benchmarks)
- `exclude_patterns`: Default file patterns to exclude from refactoring

This skill is essential for any agent that needs to improve code quality, reduce technical debt, optimize performance, or systematically refactor codebases while maintaining correctness and security.