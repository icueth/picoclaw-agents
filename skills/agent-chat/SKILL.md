---
name: agent-chat
description: Real-time chat and messaging system for AI agents with multi-platform support and collaboration features
---

# Agent Chat

This built-in skill provides real-time chat and messaging capabilities for AI agents to communicate with users, other agents, and external services across multiple platforms.

## Capabilities

- **Multi-Platform Support**: Integrate with Slack, Discord, Telegram, WhatsApp, and custom chat systems
- **Real-time Messaging**: Send and receive messages in real-time with low latency
- **Rich Media Support**: Handle text, images, files, links, and interactive elements
- **Group Chat Management**: Manage group chats, channels, and conversation threads
- **Message History**: Access and analyze message history for context awareness
- **Chat Bots Integration**: Integrate with existing chat bot frameworks and APIs
- **Presence Management**: Track online/offline status and availability
- **Message Formatting**: Support rich text formatting, markdown, and platform-specific features
- **Security Features**: Handle end-to-end encryption, secure file sharing, and access control
- **Analytics and Monitoring**: Track chat metrics, response times, and user engagement

## Usage Examples

### Send Message
```yaml
tool: agent-chat
action: send_message
platform: "slack"
channel: "#general"
message: "Hello team! Here's the latest update..."
attachments:
  - type: "image"
    url: "https://example.com/chart.png"
  - type: "file"
    path: "/reports/weekly.pdf"
```

### Read Messages
```yaml
tool: agent-chat
action: read_messages
platform: "discord"
channel: "project-updates"
limit: 50
since: "2026-03-12T00:00:00Z"
filter:
  from: ["team-member", "project-bot"]
```

### Create Channel
```yaml
tool: agent-chat
action: create_channel
platform: "slack"
name: "project-alpha"
purpose: "Coordination for Project Alpha"
members:
  - "user1@example.com"
  - "user2@example.com"
  - "agent-alpha"
```

## Security Considerations

- Message data is encrypted at rest and in transit using industry-standard protocols
- Access control ensures only authorized agents can access specific channels or messages
- Secure credential management handles authentication tokens and API keys
- Audit logging tracks all chat operations for compliance and security monitoring
- Data retention policies automatically expire sensitive message data

## Configuration

The agent-chat skill can be configured with the following parameters:

- `default_platform`: Default chat platform (default: slack)
- `message_retention`: Message retention period (default: 30 days)
- `auto_response_enabled`: Enable automated responses (default: false)
- `presence_status`: Default presence status (online, away, do_not_disturb)
- `notification_settings`: Notification preferences for different message types
- `platform_integrations`: Enabled chat platform integrations

This skill is essential for any agent that needs to communicate in real-time, collaborate with teams, or integrate with chat-based workflows. It provides seamless communication across platforms while maintaining security and privacy.