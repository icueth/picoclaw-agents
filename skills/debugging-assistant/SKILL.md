---
name: debugging-assistant
description: Intelligent debugging and error diagnosis system for AI agents with multi-language support and root cause analysis capabilities
---

# Debugging Assistant

This built-in skill provides intelligent debugging and error diagnosis capabilities for AI agents to identify, analyze, and resolve software bugs and runtime errors across multiple programming languages and environments.

## Capabilities

- **Error Log Analysis**: Parse and analyze error logs, stack traces, and crash reports to identify root causes
- **Multi-Language Debugging**: Support debugging for over 50 programming languages (Python, JavaScript, Java, Go, Rust, C++, etc.)
- **Runtime Error Diagnosis**: Diagnose runtime errors, exceptions, and crashes with detailed explanations
- **Performance Debugging**: Identify performance bottlenecks, memory leaks, and resource exhaustion issues
- **Test Case Generation**: Generate test cases to reproduce and verify bug fixes
- **Code Fix Suggestions**: Provide specific code fixes and workarounds for identified issues
- **Environment Analysis**: Analyze runtime environment, dependencies, and configuration issues
- **Debugging Workflow**: Guide users through systematic debugging workflows and troubleshooting steps
- **Historical Pattern Recognition**: Recognize patterns from historical bug reports and solutions
- **Integration with Debuggers**: Integrate with popular debuggers and development tools (VS Code, Chrome DevTools, GDB, etc.)

## Usage Examples

### Error Log Analysis
```yaml
tool: debugging-assistant
action: analyze_error_log
log_content: |
  Traceback (most recent call last):
    File "app.py", line 42, in <module>
      result = process_data(data)
    File "processor.py", line 18, in process_data
      return data['items'][0]['value']
  KeyError: 'value'
language: "python"
context:
  - "app.py"
  - "processor.py"
output_format: "detailed"
include_fixes: true
```

### Runtime Error Diagnosis
```yaml
tool: debugging-assistant
action: diagnose_runtime_error
error_type: "NullPointerException"
error_message: "Cannot read property 'length' of undefined"
stack_trace: |
  at processData (utils.js:24:32)
  at handleRequest (server.js:156:18)
  at IncomingMessage.<anonymous> (server.js:89:12)
language: "javascript"
environment:
  node_version: "18.17.0"
  dependencies:
    express: "4.18.2"
    lodash: "4.17.21"
output_format: "step_by_step"
```

### Performance Debugging
```yaml
tool: debugging-assistant
action: analyze_performance
metrics:
  cpu_usage: 95
  memory_usage: 2.1
  response_time: 4500
  error_rate: 0.05
logs:
  - "/var/log/app/performance.log"
  - "/var/log/system/memory.log"
profiling_data: "/tmp/profile.cpuprofile"
analysis_depth: "comprehensive"
recommendations: true
```

### Test Case Generation
```yaml
tool: debugging-assistant
action: generate_test_cases
bug_description: "Function fails when input array is empty"
code_snippet: |
  function processArray(items) {
    return items[0].value * 2;
  }
language: "javascript"
test_framework: "jest"
edge_cases: true
include_fixes: true
```

## Security Considerations

- Error logs and code snippets are processed in isolated environments to prevent code execution
- Sensitive information in logs is automatically redacted or filtered
- Access control ensures only authorized agents can access debugging information
- Audit logging tracks all debugging activities for compliance and security monitoring
- Generated test cases and fixes are validated for security vulnerabilities before recommendation

## Configuration

The debugging-assistant skill can be configured with the following parameters:

- `default_languages`: Default programming languages for debugging support
- `log_redaction_enabled`: Enable automatic log redaction for sensitive information (default: true)
- `max_analysis_depth`: Maximum depth of error analysis (shallow, moderate, deep)
- `test_frameworks`: Supported test frameworks (jest, pytest, junit, go test)
- `debugger_integrations`: Enabled debugger integrations (vscode, chrome, gdb, lldb)

This skill is essential for any agent that needs to debug software issues, analyze error logs, diagnose runtime problems, or provide systematic troubleshooting guidance for developers.