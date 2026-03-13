---
name: agent-mail
description: Comprehensive email management system for AI agents with multi-platform support and advanced features
---

# Agent Mail

This built-in skill provides comprehensive email management capabilities for AI agents to handle email communication, automation, and integration across multiple platforms.

## Capabilities

- **Multi-Platform Support**: Integrate with Gmail, Outlook, Fastmail, and IMAP/SMTP servers
- **Email Composition**: Compose professional emails with rich text formatting and attachments
- **Email Reading**: Read and parse incoming emails with intelligent content extraction
- **Email Filtering**: Apply intelligent filters and rules for email organization
- **Automated Responses**: Generate and send automated responses based on email content
- **Email Templates**: Manage and use email templates for consistent communication
- **Contact Management**: Integrate with contact databases for recipient management
- **Email Analytics**: Track email open rates, response times, and engagement metrics
- **Security Features**: Handle email encryption, digital signatures, and secure attachments
- **Batch Operations**: Perform batch operations on multiple emails simultaneously

## Usage Examples

### Send Email
```yaml
tool: agent-mail
action: send_email
email:
  to: ["recipient@example.com"]
  cc: ["manager@example.com"]
  subject: "Project Update"
  body: "Here's the latest update on our project..."
  attachments:
    - "/reports/project_update.pdf"
  priority: "normal"
```

### Read Emails
```yaml
tool: agent-mail
action: read_emails
filters:
  from: "important-client@example.com"
  subject_contains: "urgent"
  unread_only: true
  limit: 10
```

### Automated Response
```yaml
tool: agent-mail
action: auto_respond
trigger:
  from: "support@company.com"
  subject_contains: "ticket"
response_template: "support_acknowledgment"
variables:
  ticket_id: "{{extracted_ticket_id}}"
  expected_response_time: "24 hours"
```

## Security Considerations

- Email credentials are securely stored using encrypted credential management
- Sensitive email content is encrypted at rest and in transit
- Access control ensures only authorized agents can send or read emails
- Audit logging tracks all email operations for compliance and security
- Spam and phishing detection prevents malicious email handling

## Configuration

The agent-mail skill can be configured with the following parameters:

- `default_account`: Default email account for sending (default: primary)
- `signature`: Default email signature for outgoing messages
- `auto_reply_enabled`: Enable automated responses (default: false)
- `spam_filter_level`: Spam filtering level (low, medium, high, strict)
- `attachment_size_limit`: Maximum attachment size (default: 25MB)
- `platform_integrations`: Enabled email platform integrations

This skill is essential for any agent that needs to handle email communication, automate email workflows, or integrate with email-based systems. It provides secure and reliable email management while maintaining privacy and compliance.