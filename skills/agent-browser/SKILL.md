---
name: agent-browser
description: Headless browser automation for AI agents with comprehensive web interaction capabilities
---

# Agent Browser

This built-in skill provides headless browser automation capabilities for AI agents to interact with web pages, perform automated testing, scrape data, and handle complex web interactions.

## Capabilities

- **Page Navigation**: Navigate to URLs, handle redirects, manage history
- **Element Interaction**: Click buttons, fill forms, select options, handle dynamic content
- **Data Extraction**: Scrape text, images, tables, and structured data from web pages
- **Screenshot Capture**: Take full-page or element-specific screenshots
- **PDF Generation**: Convert web pages to PDF documents
- **JavaScript Execution**: Run custom JavaScript in the page context
- **Network Monitoring**: Intercept and analyze network requests/responses
- **Cookie Management**: Handle authentication cookies and session management
- **Wait Conditions**: Wait for elements, text, or conditions before proceeding
- **Error Handling**: Robust error handling for network issues, timeouts, and page errors

## Usage Examples

### Basic Web Scraping
```yaml
tool: agent-browser
action: scrape
url: "https://example.com"
selectors:
  - ".product-title"
  - ".price"
  - ".description"
```

### Form Automation
```yaml
tool: agent-browser
action: fill_form
url: "https://login.example.com"
fields:
  username: "{{user.username}}"
  password: "{{user.password}}"
submit: true
```

### Screenshot Capture
```yaml
tool: agent-browser
action: screenshot
url: "https://dashboard.example.com"
full_page: true
output_path: "/tmp/dashboard.png"
```

## Security Considerations

- All browser interactions run in isolated sandboxed environments
- Network requests can be restricted to specific domains if needed
- Sensitive data handling follows strict privacy guidelines
- Rate limiting prevents abuse of web services

## Configuration

The agent-browser skill can be configured with the following parameters:

- `timeout`: Maximum time to wait for page load (default: 30s)
- `headless`: Run in headless mode (default: true)
- `user_agent`: Custom user agent string
- `viewport`: Browser viewport dimensions
- `proxy`: Proxy configuration for network requests

This skill is essential for any agent that needs to interact with web-based services, perform automated testing, or extract data from websites.