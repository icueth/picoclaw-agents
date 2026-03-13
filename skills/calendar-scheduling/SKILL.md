---
name: calendar-scheduling
description: Comprehensive calendar and scheduling management for AI agents with multi-platform support
---

# Calendar Scheduling

This built-in skill provides comprehensive calendar and scheduling management capabilities for AI agents to handle appointments, meetings, reminders, and time-based coordination across multiple platforms.

## Capabilities

- **Multi-Platform Support**: Integrate with Google Calendar, Outlook, Apple Calendar, and CalDAV
- **Event Management**: Create, update, delete, and query calendar events
- **Meeting Scheduling**: Automate meeting scheduling with availability checking and conflict resolution
- **Recurring Events**: Handle recurring events with complex patterns (daily, weekly, monthly, yearly)
- **Time Zone Handling**: Automatically handle time zone conversions and daylight saving time
- **Reminders and Notifications**: Set up reminders and notifications for upcoming events
- **Calendar Sharing**: Manage calendar sharing permissions and access control
- **Busy Time Detection**: Detect busy times and suggest optimal meeting slots
- **Integration with Contacts**: Link events with contact information for easy coordination
- **Export/Import**: Export calendar data to various formats and import from external sources

## Usage Examples

### Create Event
```yaml
tool: calendar-scheduling
action: create_event
event:
  title: "Team Meeting"
  description: "Weekly team sync meeting"
  start_time: "2026-03-15T14:00:00Z"
  end_time: "2026-03-15T15:00:00Z"
  timezone: "America/New_York"
  attendees:
    - "john@example.com"
    - "jane@example.com"
  calendar: "work"
```

### Find Available Time
```yaml
tool: calendar-scheduling
action: find_available_time
duration: "60m"
within_days: 7
preferred_hours:
  start: "09:00"
  end: "17:00"
attendees:
  - "user@example.com"
  - "colleague@example.com"
```

### Schedule Meeting
```yaml
tool: calendar-scheduling
action: schedule_meeting
meeting:
  title: "Project Review"
  duration: "90m"
  attendees:
    - "user@example.com"
    - "manager@example.com"
    - "team@example.com"
  constraints:
    - "weekdays_only"
    - "business_hours"
  send_invites: true
```

## Security Considerations

- Calendar data is encrypted at rest and in transit
- Access control ensures only authorized agents can modify calendar events
- Privacy settings respect user preferences for event visibility
- Audit logging tracks all calendar modifications for compliance
- Secure authentication handles OAuth tokens and credentials safely

## Configuration

The calendar-scheduling skill can be configured with the following parameters:

- `default_calendar`: Default calendar for new events (default: primary)
- `time_zone`: Default time zone for events (default: system time zone)
- `notification_methods`: Notification methods (email, push, internal)
- `availability_window`: Days ahead to consider for scheduling (default: 30)
- `business_hours`: Default business hours for scheduling (default: 9-5)
- `platform_integrations`: Enabled calendar platform integrations

This skill is essential for any agent that needs to manage time-based activities, coordinate with others, or handle scheduling tasks. It provides seamless integration with popular calendar platforms while maintaining security and privacy.