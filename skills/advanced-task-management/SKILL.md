---
name: advanced-task-management
description: Intelligent task management and productivity system for AI agents with advanced planning, prioritization, and collaboration capabilities
---

# Advanced Task Management

This built-in skill provides intelligent task management and productivity capabilities for AI agents to organize, prioritize, and execute complex workflows with advanced planning, collaboration, and time management features.

## Capabilities

- **Task Organization**: Create, organize, and manage tasks with hierarchical structures, tags, and custom fields
- **Intelligent Prioritization**: Automatically prioritize tasks based on deadlines, importance, dependencies, and energy levels
- **Time Blocking**: Implement time blocking strategies with optimal scheduling based on user preferences and constraints
- **Dependency Management**: Manage task dependencies and ensure proper execution order with automatic rescheduling
- **Collaboration Features**: Enable team collaboration with shared tasks, assignments, and progress tracking
- **Progress Tracking**: Track task completion progress with percentage completion, time logging, and milestone tracking
- **Habit Formation**: Support habit formation and routine establishment with streak tracking and reminders
- **Integration Hub**: Integrate with popular task management tools (Todoist, Trello, Asana, Jira, Microsoft To Do)
- **Analytics and Insights**: Provide productivity analytics and insights with trend analysis and improvement recommendations
- **Context Switching Minimization**: Minimize context switching by grouping similar tasks and optimizing workflow sequences

## Usage Examples

### Intelligent Task Planning
```yaml
tool: advanced-task-management
action: create_task_plan
plan_name: "Q2 Project Delivery"
tasks:
  - name: "Requirements Gathering"
    description: "Gather requirements from stakeholders"
    priority: "high"
    deadline: "2026-04-15"
    estimated_duration: "8h"
    energy_level: "high"
    tags: ["planning", "stakeholder"]
  - name: "System Design"
    description: "Create system architecture and design documents"
    priority: "high"
    deadline: "2026-04-30"
    estimated_duration: "16h"
    energy_level: "high"
    dependencies: ["Requirements Gathering"]
    tags: ["design", "architecture"]
  - name: "Implementation"
    description: "Implement core features and functionality"
    priority: "medium"
    deadline: "2026-05-31"
    estimated_duration: "40h"
    energy_level: "medium"
    dependencies: ["System Design"]
    tags: ["development", "coding"]
time_blocking:
  high_energy: ["09:00-11:00", "14:00-16:00"]
  medium_energy: ["11:00-12:00", "16:00-17:00"]
  low_energy: ["13:00-14:00", "17:00-18:00"]
```

### Team Collaboration
```yaml
tool: advanced-task-management
action: manage_team_tasks
project: "Marketing Campaign Q2"
team_members:
  - name: "Sarah"
    role: "Content Creator"
    tasks: ["Blog Posts", "Social Media Content"]
  - name: "Mike"
    role: "Designer"
    tasks: ["Banner Design", "Email Templates"]
  - name: "John"
    role: "Project Manager"
    tasks: ["Timeline Management", "Stakeholder Communication"]
collaboration_features:
  - "shared_progress_tracking"
  - "automatic_notifications"
  - "dependency_visualization"
  - "workload_balancing"
integration: "asana"
```

### Habit Formation
```yaml
tool: advanced-task-management
action: create_habit_system
habits:
  - name: "Morning Exercise"
    frequency: "daily"
    time: "07:00"
    duration: "30m"
    streak_tracking: true
    reminders: ["06:45", "06:55"]
  - name: "Weekly Planning"
    frequency: "weekly"
    day: "sunday"
    time: "19:00"
    duration: "60m"
    streak_tracking: true
    reminders: ["sunday_18:30"]
  - name: "Daily Review"
    frequency: "daily"
    time: "20:00"
    duration: "15m"
    streak_tracking: true
    reminders: ["19:45"]
motivation_strategies:
  - "celebration_on_milestones"
  - "progress_visualization"
  - "accountability_partners"
```

### Productivity Analytics
```yaml
tool: advanced-task-management
action: analyze_productivity
time_period: "last_30_days"
metrics:
  - "task_completion_rate"
  - "time_spent_by_category"
  - "energy_level_correlation"
  - "context_switching_frequency"
  - "deadline_adherence"
insights:
  - "most_productive_time_blocks"
  - "task_category_performance"
  - "improvement_recommendations"
  - "habit_consistency_analysis"
report_format: "interactive_dashboard"
```

## Security Considerations

- Task data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific task information
- Privacy controls allow users to manage what task data is shared and with whom
- Audit logging tracks all task management activities for accountability and security
- Integration credentials are securely stored and never exposed in logs

## Configuration

The advanced-task-management skill can be configured with the following parameters:

- `default_priority_method`: Default task prioritization method (deadline, importance, energy_match)
- `time_blocking_enabled`: Enable time blocking by default (default: true)
- `collaboration_features`: Enabled collaboration features (sharing, assignments, notifications)
- `analytics_retention`: Analytics data retention period (default: 90 days)
- `privacy_level`: Privacy level for task data sharing (private, team, public)

This skill is essential for any agent that needs to manage complex workflows, optimize productivity, enable team collaboration, or provide intelligent task management and planning capabilities.