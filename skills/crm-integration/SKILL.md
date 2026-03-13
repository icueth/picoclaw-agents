---
name: crm-integration
description: Comprehensive CRM integration and management system for AI agents with multi-platform support and workflow automation capabilities
---

# CRM Integration

This built-in skill provides comprehensive CRM integration and management capabilities for AI agents to connect with, manage, and automate workflows across multiple CRM platforms and customer data systems.

## Capabilities

- **Multi-Platform Support**: Integrate with popular CRM platforms (Salesforce, HubSpot, Pipedrive, Zoho, Microsoft Dynamics)
- **Contact Management**: Create, update, and manage contacts, leads, and accounts with comprehensive data fields
- **Deal Pipeline Management**: Manage sales pipelines, deals, opportunities, and stages with custom workflows
- **Activity Tracking**: Track and log activities (calls, emails, meetings, tasks) with automatic synchronization
- **Data Synchronization**: Synchronize data between CRM and external systems with conflict resolution
- **Workflow Automation**: Automate CRM workflows based on triggers, conditions, and actions
- **Reporting and Analytics**: Generate CRM reports and analytics on sales performance, pipeline health, and team metrics
- **Custom Object Management**: Manage custom objects and fields specific to business requirements
- **API Integration**: Connect CRM data with external APIs and services for extended functionality
- **Data Enrichment**: Enrich CRM records with additional data from external sources and databases

## Usage Examples

### Create Contact
```yaml
tool: crm-integration
action: create_contact
crm: "hubspot"
contact:
  first_name: "John"
  last_name: "Doe"
  email: "john.doe@example.com"
  phone: "+1-555-123-4567"
  company: "Tech Corp"
  job_title: "CTO"
  lifecycle_stage: "lead"
  custom_properties:
    tech_stack: "python, kubernetes, aws"
    company_size: "200"
    industry: "fintech"
```

### Manage Deal Pipeline
```yaml
tool: crm-integration
action: update_deal
crm: "pipedrive"
deal:
  id: "deal-12345"
  title: "Enterprise AI Platform - Tech Corp"
  value: 125000
  currency: "USD"
  stage: "proposal_sent"
  expected_close_date: "2026-04-15"
  probability: 75
  custom_fields:
    product_tier: "enterprise"
    implementation_timeline: "3_months"
    decision_maker: "john.doe@example.com"
```

### Workflow Automation
```yaml
tool: crm-integration
action: create_workflow
crm: "salesforce"
workflow:
  name: "Lead Qualification Workflow"
  trigger:
    object: "lead"
    event: "created"
    conditions:
      - field: "lead_score"
        operator: "greater_than"
        value: 75
  actions:
    - type: "assign_owner"
      owner: "sales_team"
    - type: "send_email"
      template: "lead_welcome"
      delay: "0h"
    - type: "create_task"
      subject: "Follow up with qualified lead"
      due_date: "+2d"
      priority: "high"
    - type: "update_field"
      field: "status"
      value: "qualified"
```

### Data Synchronization
```yaml
tool: crm-integration
action: sync_data
source_crm: "hubspot"
target_crm: "salesforce"
sync_direction: "bidirectional"
objects:
  - "contacts"
  - "companies"
  - "deals"
field_mapping:
  hubspot.email: salesforce.email
  hubspot.company: salesforce.account_name
  hubspot.lifecycle_stage: salesforce.lead_status
conflict_resolution: "newest_wins"
schedule: "hourly"
```

## Security Considerations

- CRM data is encrypted at rest and in transit using industry-standard encryption
- API credentials are securely managed using encrypted credential storage
- Access control ensures only authorized agents can access specific CRM data and functions
- Audit logging tracks all CRM operations for compliance and security monitoring
- Data synchronization includes conflict resolution and data integrity validation

## Configuration

The crm-integration skill can be configured with the following parameters:

- `default_crm`: Default CRM platform (hubspot, salesforce, pipedrive, zoho)
- `sync_frequency`: Default synchronization frequency (real_time, hourly, daily)
- `field_mapping`: Default field mappings between different CRM platforms
- `conflict_resolution`: Default conflict resolution strategy (newest_wins, source_wins, manual)
- `audit_logging_level`: Audit logging level for CRM operations (minimal, standard, comprehensive)

This skill is essential for any agent that needs to manage customer relationships, automate sales workflows, synchronize data across platforms, or integrate CRM functionality into broader business processes.