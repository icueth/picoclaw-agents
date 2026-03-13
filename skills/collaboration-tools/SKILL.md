---
name: collaboration-tools
description: Comprehensive collaboration and team communication system for AI agents with multi-platform support and intelligent workflow integration capabilities
---

# Collaboration Tools

This built-in skill provides comprehensive collaboration and team communication capabilities for AI agents to facilitate effective teamwork, knowledge sharing, and project coordination across multiple platforms and communication channels.

## Capabilities

- **Multi-Platform Support**: Integrate with popular collaboration platforms (Slack, Microsoft Teams, Discord, Zoom, Google Workspace, Notion, Confluence)
- **Intelligent Communication**: Facilitate intelligent communication with context-aware responses, summaries, and action item extraction
- **Document Collaboration**: Enable real-time document collaboration with version control, commenting, and approval workflows
- **Meeting Management**: Manage meetings with automated scheduling, agendas, note-taking, and follow-up tracking
- **Knowledge Management**: Organize and share team knowledge with searchable repositories and intelligent retrieval
- **Project Coordination**: Coordinate project activities with task assignment, progress tracking, and dependency management
- **File Sharing**: Manage file sharing with version control, access controls, and automatic organization
- **Notification Management**: Manage notifications intelligently to prevent information overload while ensuring important updates are delivered
- **Integration Hub**: Connect collaboration tools with other productivity systems (task management, calendar, email)
- **Analytics and Insights**: Provide collaboration analytics and insights for team performance optimization

## Usage Examples

### Multi-Platform Communication
```yaml
tool: collaboration-tools
action: send_message
platforms:
  - name: "slack"
    channel: "#project-alpha"
    message: "Weekly project update: We've completed the design phase and are moving to implementation."
    attachments:
      - type: "file"
        path: "/reports/weekly_update.pdf"
      - type: "link"
        url: "https://docs.example.com/project-alpha"
  - name: "teams"
    channel: "Project Alpha"
    message: "Weekly project update: We've completed the design phase and are moving to implementation."
    mentions: ["@john.doe", "@sarah.smith"]
priority: "normal"
```

### Meeting Management
```yaml
tool: collaboration-tools
action: manage_meeting
meeting:
  title: "Project Alpha Weekly Sync"
  date: "2026-03-15"
  time: "14:00"
  duration: "60m"
  platform: "zoom"
  attendees:
    - "john@example.com"
    - "sarah@example.com"
    - "mike@example.com"
agenda:
  - "Review last week's accomplishments"
  - "Discuss current blockers"
  - "Plan next week's priorities"
  - "Assign action items"
automated_features:
  - "transcription"
  - "note_taking"
  - "action_item_extraction"
  - "follow_up_tracking"
calendar_integration: true
```

### Knowledge Management
```yaml
tool: collaboration-tools
action: manage_knowledge
repository: "Project Alpha Knowledge Base"
documents:
  - title: "System Architecture Overview"
    content_path: "/docs/architecture.md"
    tags: ["architecture", "design", "technical"]
    access_level: "team"
  - title: "API Documentation"
    content_path: "/docs/api.md"
    tags: ["api", "documentation", "technical"]
    access_level: "team"
  - title: "User Research Findings"
    content_path: "/docs/research.md"
    tags: ["research", "user_experience", "design"]
    access_level: "team"
search_enabled: true
version_control: true
approval_workflow: true
```

### Project Coordination
```yaml
tool: collaboration-tools
action: coordinate_project
project: "Project Alpha"
coordination_features:
  - "task_assignment"
  - "progress_tracking"
  - "dependency_management"
  - "resource_allocation"
  - "risk_management"
  - "stakeholder_communication"
integrations:
  - platform: "jira"
    project_key: "PA"
  - platform: "confluence"
    space_key: "PA"
  - platform: "slack"
    channel: "#project-alpha"
notification_rules:
  - event: "task_completed"
    recipients: ["project_manager", "team_lead"]
  - event: "deadline_approaching"
    recipients: ["task_assignee", "project_manager"]
  - event: "blocker_identified"
    recipients: ["entire_team", "stakeholders"]
```

## Security Considerations

- Collaboration data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific collaboration information
- Privacy controls allow users to manage what collaboration data is shared and with whom
- Audit logging tracks all collaboration activities for accountability and security
- Integration credentials are securely stored and never exposed in logs

## Configuration

The collaboration-tools skill can be configured with the following parameters:

- `default_platforms`: Default collaboration platforms (slack, teams, zoom, google_workspace)
- `notification_level`: Notification level (minimal, standard, comprehensive)
- `knowledge_access_level`: Default knowledge access level (private, team, organization)
- `meeting_automation_enabled`: Enable meeting automation features (default: true)
- `privacy_level`: Privacy level for collaboration data sharing (private, team, public)

This skill is essential for any agent that needs to facilitate team collaboration, manage communication across platforms, coordinate project activities, or provide intelligent collaboration and knowledge management capabilities.