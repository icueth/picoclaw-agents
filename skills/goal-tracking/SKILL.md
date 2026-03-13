---
name: goal-tracking
description: Intelligent goal setting and progress tracking system for AI agents with OKR support, milestone management, and achievement analytics capabilities
---

# Goal Tracking

This built-in skill provides intelligent goal setting and progress tracking capabilities for AI agents to help users define, pursue, and achieve their personal and professional objectives through structured frameworks and continuous monitoring.

## Capabilities

- **Goal Framework Support**: Support multiple goal frameworks (SMART goals, OKRs, BHAGs, WOOP) with proper structure and validation
- **Milestone Management**: Break down goals into achievable milestones with clear success criteria and timelines
- **Progress Tracking**: Track goal progress with quantitative metrics, qualitative assessments, and visual indicators
- **Habit Integration**: Connect goals to daily habits and routines for consistent progress toward long-term objectives
- **Accountability Systems**: Implement accountability mechanisms with regular check-ins, progress reviews, and stakeholder updates
- **Obstacle Management**: Identify and address obstacles that prevent goal achievement with adaptive strategies
- **Motivation Enhancement**: Provide motivation enhancement through celebration of wins, visualization of progress, and reminder systems
- **Integration with Tasks**: Connect goals to daily tasks and time management systems for actionable execution
- **Analytics and Insights**: Provide goal achievement analytics with trend analysis, correlation insights, and improvement recommendations
- **Collaborative Goals**: Support team and collaborative goals with shared ownership, progress tracking, and coordination

## Usage Examples

### OKR Framework Implementation
```yaml
tool: goal-tracking
action: create_okr
objective:
  title: "Launch Product Alpha by Q2 2026"
  description: "Successfully launch our product alpha to early adopters with core features complete"
  timeline: "Q2 2026"
key_results:
  - title: "Complete core feature development"
    metric: "features_completed"
    target: 15
    current: 8
    confidence: "high"
  - title: "Achieve 95% test coverage"
    metric: "test_coverage_percentage"
    target: 95
    current: 78
    confidence: "medium"
  - title: "Onboard 50 beta testers"
    metric: "beta_testers_count"
    target: 50
    current: 12
    confidence: "low"
  - title: "Maintain system uptime of 99.9%"
    metric: "system_uptime_percentage"
    target: 99.9
    current: 99.5
    confidence: "high"
review_frequency: "weekly"
stakeholders: ["product_team", "engineering_team", "executives"]
```

### SMART Goal Creation
```yaml
tool: goal-tracking
action: create_smart_goal
goal:
  specific: "Increase monthly active users from 1,000 to 5,000"
  measurable: "Track MAU through analytics dashboard"
  achievable: "Based on current growth rate and marketing budget"
  relevant: "Aligns with company objective to expand user base"
  time_bound: "By December 31, 2026"
milestones:
  - title: "Implement referral program"
    deadline: "2026-04-30"
    success_criteria: "Referral program live with 100+ referrals"
  - title: "Launch social media campaign"
    deadline: "2026-06-30"
    success_criteria: "Campaign reaches 100,000 impressions"
  - title: "Optimize onboarding flow"
    deadline: "2026-08-31"
    success_criteria: "Onboarding completion rate increases to 80%"
habits:
  - name: "Daily user acquisition review"
    frequency: "daily"
    duration: "15m"
  - name: "Weekly growth metric analysis"
    frequency: "weekly"
    duration: "60m"
```

### Progress Tracking and Analytics
```yaml
tool: goal-tracking
action: track_progress
goal_id: "okr-q2-2026"
metrics:
  - name: "features_completed"
    current_value: 12
    target_value: 15
    last_updated: "2026-03-13"
  - name: "test_coverage_percentage"
    current_value: 85
    target_value: 95
    last_updated: "2026-03-13"
  - name: "beta_testers_count"
    current_value: 25
    target_value: 50
    last_updated: "2026-03-13"
insights:
  - "Feature development is on track with 80% completion"
  - "Test coverage improving steadily, currently at 85%"
  - "Beta tester acquisition needs acceleration, only at 50% of target"
  - "Overall confidence in Q2 delivery: 75%"
visualization: "progress_dashboard"
```

### Obstacle Management
```yaml
tool: goal-tracking
action: manage_obstacles
goal_id: "okr-q2-2026"
obstacles:
  - description: "Engineering team bandwidth constraints"
    impact: "high"
    mitigation_strategy: "Hire temporary contractor for 2 months"
    status: "in_progress"
  - description: "Beta tester recruitment slower than expected"
    impact: "medium"
    mitigation_strategy: "Increase referral incentives and partner with tech communities"
    status: "planned"
  - description: "Third-party API reliability issues"
    impact: "medium"
    mitigation_strategy: "Implement caching layer and fallback mechanisms"
    status: "completed"
risk_assessment: "medium"
contingency_plan: "Extend beta period by 2 weeks if needed"
```

## Security Considerations

- Goal data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific goal information
- Privacy controls allow users to manage what goal data is shared and with whom
- Audit logging tracks all goal tracking activities for accountability and security
- Sensitive performance data is protected with appropriate access controls

## Configuration

The goal-tracking skill can be configured with the following parameters:

- `default_framework`: Default goal framework (smart, okr, bhag, woop)
- `review_frequency`: Default review frequency (daily, weekly, monthly, quarterly)
- `analytics_retention`: Analytics data retention period (default: 365 days)
- `collaboration_enabled`: Enable collaborative goals by default (default: true)
- `privacy_level`: Privacy level for goal data sharing (private, team, organization)

This skill is essential for any agent that needs to help users set and achieve meaningful goals, track progress systematically, overcome obstacles, or provide structured goal management and achievement capabilities.