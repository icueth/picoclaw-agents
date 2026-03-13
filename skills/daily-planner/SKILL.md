---
name: daily-planner
description: Intelligent daily planning and productivity system for AI agents with task prioritization and time management features
---

# Daily Planner

This built-in skill provides intelligent daily planning and productivity capabilities for AI agents to organize tasks, manage time effectively, and optimize daily workflows based on priorities and constraints.

## Capabilities

- **Task Prioritization**: Automatically prioritize tasks based on deadlines, importance, and dependencies
- **Time Blocking**: Allocate specific time blocks for different types of activities and tasks
- **Energy Management**: Schedule tasks based on user's energy levels and peak performance times
- **Goal Alignment**: Ensure daily activities align with long-term goals and objectives
- **Context Switching**: Minimize context switching by grouping similar tasks together
- **Buffer Time**: Automatically allocate buffer time for unexpected interruptions and task overruns
- **Progress Tracking**: Track completion progress and adjust plans dynamically throughout the day
- **Review and Reflection**: Conduct end-of-day reviews to capture learnings and improve future planning
- **Integration with Calendar**: Sync with calendar systems to avoid scheduling conflicts
- **Habit Formation**: Support habit formation and routine establishment for consistent productivity

## Usage Examples

### Create Daily Plan
```yaml
tool: daily-planner
action: create_daily_plan
date: "2026-03-14"
constraints:
  available_hours:
    start: "09:00"
    end: "18:00"
  breaks:
    - start: "12:00"
      duration: "60m"
    - start: "15:00"
      duration: "15m"
priorities:
  - "Complete project deadline"
  - "Team meeting preparation"
  - "Email responses"
  - "Learning new skill"
energy_levels:
  high: ["09:00", "14:00"]
  medium: ["11:00", "16:00"]
  low: ["10:00", "17:00"]
```

### Update Plan
```yaml
tool: daily-planner
action: update_plan
plan_id: "plan-2026-03-14"
updates:
  completed_tasks:
    - "Morning email review"
  rescheduled_tasks:
    - task: "Project work"
      new_time: "14:00-16:00"
  new_tasks:
    - name: "Urgent client call"
      priority: "high"
      duration: "30m"
```

### End-of-Day Review
```yaml
tool: daily-planner
action: end_of_day_review
plan_id: "plan-2026-03-14"
actual_completion:
  completed: 8
  total: 12
  interrupted_by: ["urgent_meeting", "system_outage"]
learnings:
  - "Need more buffer time for coding tasks"
  - "Morning is better for deep work"
improvements:
  - "Schedule deep work in morning energy peak"
  - "Add 15-minute buffer between meetings"
```

## Security Considerations

- Personal productivity data is encrypted at rest and handled with privacy considerations
- Access control ensures only authorized agents can access or modify personal plans
- Data minimization principles limit collection to essential planning information
- Audit logging tracks plan modifications for accountability and improvement
- Secure integration with calendar systems protects scheduling privacy

## Configuration

The daily-planner skill can be configured with the following parameters:

- `default_work_hours`: Default working hours (default: 9-5)
- `time_block_size`: Default time block size (default: 90 minutes)
- `buffer_percentage`: Percentage of time allocated for buffers (default: 20%)
- `priority_method`: Task prioritization method (deadline, importance, energy_match)
- `review_frequency`: Frequency of plan reviews (daily, weekly, monthly)
- `integration_enabled`: Enable calendar and task management integrations

This skill is essential for any agent that needs to help users manage their time effectively, stay organized, and achieve their daily goals. It provides intelligent planning capabilities while respecting personal preferences and constraints.