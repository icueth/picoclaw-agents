---
name: agent-contact-card
description: Contact management and networking system for AI agents with vCard support and relationship tracking
---

# Agent Contact Card

This built-in skill provides contact management and networking capabilities for AI agents to store, organize, and manage contact information with relationship tracking and integration features.

## Capabilities

- **Contact Storage**: Store comprehensive contact information with structured fields
- **vCard Support**: Import/export contacts using vCard format (v3.0, v4.0)
- **Relationship Tracking**: Track relationships, interactions, and communication history
- **Contact Grouping**: Organize contacts into groups, categories, and tags
- **Social Integration**: Integrate with social networks and professional platforms
- **Contact Enrichment**: Automatically enrich contact data from public sources
- **Privacy Management**: Control privacy settings and data sharing permissions
- **Search and Discovery**: Search contacts using natural language queries and filters
- **Contact Validation**: Validate and verify contact information accuracy
- **Export/Import**: Export contacts to various formats and import from external sources

## Usage Examples

### Create Contact
```yaml
tool: agent-contact-card
action: create_contact
contact:
  name: "John Doe"
  email: "john.doe@example.com"
  phone: "+1-555-123-4567"
  company: "Tech Corp"
  title: "Software Engineer"
  address:
    street: "123 Main St"
    city: "San Francisco"
    state: "CA"
    zip: "94105"
    country: "USA"
  tags: ["colleague", "developer", "project-alpha"]
```

### Find Contact
```yaml
tool: agent-contact-card
action: find_contact
query: "software engineer at Tech Corp"
fields: ["name", "email", "company", "title"]
limit: 5
```

### Import vCard
```yaml
tool: agent-contact-card
action: import_vcard
file_path: "/contacts/team.vcf"
format: "v4.0"
group: "project-team"
```

## Security Considerations

- Contact data is encrypted at rest using industry-standard encryption
- Privacy settings control what information is shared and with whom
- Access control ensures only authorized agents can access specific contacts
- Audit logging tracks all contact operations for compliance and security
- Data retention policies automatically expire outdated contact information

## Configuration

The agent-contact-card skill can be configured with the following parameters:

- `storage_backend`: Storage backend (sqlite, postgres, memory)
- `auto_enrichment`: Enable automatic contact enrichment (default: false)
- `privacy_level`: Default privacy level for new contacts (public, private, restricted)
- `validation_enabled`: Enable contact validation (default: true)
- `sync_frequency`: Frequency for syncing with external sources (default: daily)
- `export_formats`: Supported export formats (vcard, csv, json)

This skill is essential for any agent that needs to manage professional relationships, coordinate with contacts, or maintain organized contact information. It provides secure and efficient contact management while respecting privacy and data protection requirements.