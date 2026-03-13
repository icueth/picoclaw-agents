---
name: ide-integration
description: Comprehensive IDE integration and development environment management system for AI agents with multi-editor support and intelligent assistance capabilities
---

# IDE Integration

This built-in skill provides comprehensive IDE integration and development environment management capabilities for AI agents to interact with, control, and enhance popular integrated development environments and code editors.

## Capabilities

- **Multi-Editor Support**: Integrate with popular IDEs and editors (VS Code, Vim, Emacs, IntelliJ, Sublime Text, Atom)
- **Code Navigation**: Navigate codebases with intelligent search, go-to-definition, and find-references
- **Code Completion**: Provide intelligent code completion and suggestions based on context and patterns
- **Refactoring Tools**: Execute automated refactoring operations (rename, extract method, move class, etc.)
- **Debugging Integration**: Control debuggers and inspect runtime state within the IDE environment
- **Version Control**: Manage Git operations and version control directly from the IDE context
- **Terminal Integration**: Execute terminal commands and scripts within the IDE environment
- **Extension Management**: Install, configure, and manage IDE extensions and plugins
- **Workspace Management**: Manage IDE workspaces, projects, and configuration settings
- **Intelligent Assistance**: Provide context-aware assistance and automation within the development workflow

## Usage Examples

### Code Navigation
```yaml
tool: ide-integration
action: navigate_code
ide: "vscode"
navigation_type: "go_to_definition"
file_path: "/src/main.py"
line_number: 42
column_number: 15
workspace_root: "/project"
output_format: "location"
```

### Intelligent Refactoring
```yaml
tool: ide-integration
action: refactor_code
ide: "intellij"
refactor_type: "extract_method"
file_path: "/src/processor.java"
start_line: 25
end_line: 35
new_method_name: "validateInputData"
parameters:
  - "data"
  - "schema"
return_type: "boolean"
preview_changes: true
```

### Debugging Control
```yaml
tool: ide-integration
action: control_debugger
ide: "vscode"
debug_operation: "set_breakpoint"
file_path: "/src/app.js"
line_number: 78
condition: "user.id === 'test123'"
hit_count: 1
```

### Version Control Operations
```yaml
tool: ide-integration
action: git_operations
ide: "vim"
operations:
  - type: "add"
    paths: ["/src/new_feature.py"]
  - type: "commit"
    message: "Add new feature implementation"
    author: "AI Agent <ai@company.com>"
  - type: "push"
    remote: "origin"
    branch: "feature/new-feature"
```

### Workspace Configuration
```yaml
tool: ide-integration
action: configure_workspace
ide: "vscode"
settings:
  editor.tabSize: 2
  editor.insertSpaces: true
  files.exclude:
    "**/*.log": true
    "**/node_modules": true
extensions:
  - "ms-python.python"
  - "ms-vscode.vscode-typescript-next"
  - "github.copilot"
launch_configurations:
  - name: "Debug Application"
    type: "node"
    request: "launch"
    program: "${workspaceFolder}/src/app.js"
```

## Security Considerations

- IDE integration operations are validated against security policies before execution
- File system access is restricted to authorized project directories only
- Sensitive configuration data is encrypted and securely managed
- Access control ensures only authorized agents can modify IDE settings or execute operations
- Audit logging tracks all IDE integration activities for compliance and security monitoring

## Configuration

The ide-integration skill can be configured with the following parameters:

- `default_ide`: Default IDE/editor (vscode, vim, intellij, emacs)
- `max_file_size`: Maximum file size for operations (default: 10MB)
- `allowed_operations`: Allowed IDE operations (navigation, refactoring, debugging, git)
- `security_level`: Security level for IDE operations (strict, moderate, relaxed)
- `extension_whitelist`: Whitelist of approved IDE extensions

This skill is essential for any agent that needs to interact with development environments, automate coding tasks, provide intelligent assistance, or integrate development workflows across multiple IDEs and editors.