---
name: lead-generation
description: Intelligent lead generation and prospecting system for AI agents with multi-channel outreach and qualification capabilities
---

# Lead Generation

This built-in skill provides intelligent lead generation and prospecting capabilities for AI agents to identify, qualify, and engage potential customers across multiple channels and platforms.

## Capabilities

- **Prospect Discovery**: Discover potential leads from various sources (LinkedIn, GitHub, company websites, directories)
- **Lead Qualification**: Qualify leads based on firmographic, demographic, and behavioral criteria
- **Multi-Channel Outreach**: Execute personalized outreach campaigns across email, social media, and messaging platforms
- **Contact Enrichment**: Enrich lead data with additional information from public sources and databases
- **Campaign Management**: Create, manage, and optimize lead generation campaigns with A/B testing
- **Response Tracking**: Track and analyze lead responses and engagement metrics
- **CRM Integration**: Integrate with popular CRM systems for seamless lead management
- **Compliance Management**: Ensure compliance with data protection regulations (GDPR, CCPA, CAN-SPAM)
- **Analytics and Reporting**: Generate comprehensive analytics and reports on lead generation performance
- **Automated Follow-up**: Schedule and execute automated follow-up sequences based on lead behavior

## Usage Examples

### Prospect Discovery
```yaml
tool: lead-generation
action: discover_prospects
sources:
  - "linkedin"
  - "github"
  - "company_websites"
filters:
  company_size: "50-500"
  industry: ["technology", "software", "saas"]
  location: ["United States", "Canada"]
  job_title: ["CTO", "Engineering Manager", "DevOps Engineer"]
max_results: 100
```

### Lead Qualification
```yaml
tool: lead-generation
action: qualify_leads
leads:
  - name: "John Doe"
    company: "Tech Corp"
    title: "CTO"
    email: "john@techcorp.com"
qualification_criteria:
  budget: "verified"
  authority: "decision_maker"
  need: "matches_product"
  timeline: "within_3_months"
scoring_threshold: 75
```

### Multi-Channel Outreach
```yaml
tool: lead-generation
action: execute_outreach
campaign:
  name: "Q1 Product Launch"
  channels:
    - type: "email"
      template: "product_launch_email"
      subject: "Introducing Our New AI Platform"
    - type: "linkedin"
      template: "connection_request"
      message: "Hi {{first_name}}, I'd love to connect and share our new AI platform"
  personalization:
    company_name: "{{company}}"
    pain_point: "{{industry_pain_point}}"
    value_proposition: "{{relevant_benefit}}"
schedule:
  type: "staggered"
  interval: "2h"
  max_per_hour: 10
```

### CRM Integration
```yaml
tool: lead-generation
action: sync_to_crm
crm: "hubspot"
leads:
  - name: "Jane Smith"
    company: "Innovate Inc"
    title: "VP Engineering"
    email: "jane@innovate.com"
    phone: "+1-555-123-4567"
    source: "linkedin_outreach"
    campaign: "Q1_Product_Launch"
    status: "new_lead"
    custom_fields:
      tech_stack: "python, kubernetes, aws"
      company_size: "200"
      industry: "fintech"
```

## Security Considerations

- Lead data is encrypted at rest and in transit using industry-standard encryption
- Compliance with data protection regulations is enforced through automated checks
- Access control ensures only authorized agents can access sensitive lead information
- Audit logging tracks all lead generation activities for compliance and security
- Data retention policies automatically expire outdated lead information

## Configuration

The lead-generation skill can be configured with the following parameters:

- `default_sources`: Default prospect discovery sources (linkedin, github, company_websites)
- `compliance_regions`: Enabled compliance regions (gdpr, ccpa, can_spam)
- `max_daily_outreach`: Maximum daily outreach limit per channel (default: 100)
- `personalization_level`: Personalization level (basic, moderate, advanced)
- `crm_integrations`: Enabled CRM integrations (hubspot, salesforce, pipedrive)

This skill is essential for any agent that needs to generate qualified leads, execute targeted outreach campaigns, manage prospect relationships, or integrate with sales and marketing workflows.