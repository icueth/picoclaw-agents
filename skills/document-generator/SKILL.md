---
name: document-generator
description: Intelligent document generation system for AI agents with template support and multi-format output
---

# Document Generator

This built-in skill provides intelligent document generation capabilities for AI agents to create professional documents, reports, contracts, and other structured content from templates and data sources.

## Capabilities

- **Template Management**: Create, manage, and use document templates with dynamic variables
- **Multi-Format Output**: Generate documents in PDF, DOCX, HTML, Markdown, and plain text formats
- **Data Integration**: Populate templates with data from various sources (JSON, CSV, databases, APIs)
- **Conditional Logic**: Apply conditional formatting and content based on data values
- **Table Generation**: Create formatted tables with automatic column sizing and styling
- **Image Embedding**: Embed images, charts, and graphics into generated documents
- **Header/Footer Management**: Customize headers, footers, page numbers, and margins
- **Style Customization**: Apply custom styling, fonts, colors, and layouts
- **Batch Generation**: Generate multiple documents from a single template with different data sets
- **Version Control**: Track document versions and maintain change history

## Usage Examples

### Generate Report from Template
```yaml
tool: document-generator
action: generate_document
template:
  path: "/templates/monthly_report.docx"
  format: "docx"
data:
  month: "March 2026"
  metrics:
    revenue: "$125,000"
    users: "1,250"
    growth: "15%"
  highlights:
    - "Launched new feature X"
    - "Expanded to new market Y"
output_path: "/reports/march_2026_report.docx"
```

### Create Contract from JSON Data
```yaml
tool: document-generator
action: generate_document
template:
  type: "inline"
  content: |
    CONTRACT AGREEMENT

    This agreement is made between {{client_name}} and {{company_name}} on {{date}}.

    {% if contract_type == "service" %}
    Services to be provided: {{services}}
    {% endif %}

    Total Amount: {{amount}}
    Payment Terms: {{payment_terms}}
  format: "pdf"
data:
  client_name: "ABC Corporation"
  company_name: "Tech Solutions Inc."
  date: "2026-03-13"
  contract_type: "service"
  services: "Software development and maintenance"
  amount: "$50,000"
  payment_terms: "Net 30"
output_path: "/contracts/abc_corp_contract.pdf"
```

### Batch Generate Invoices
```yaml
tool: document-generator
action: batch_generate
template:
  path: "/templates/invoice.html"
  format: "pdf"
data_source:
  type: "csv"
  path: "/data/invoices_march.csv"
output_directory: "/invoices/march_2026/"
naming_pattern: "invoice_{{customer_id}}.pdf"
```

## Security Considerations

- Template sanitization prevents code injection and malicious content
- Data validation ensures only safe content is included in generated documents
- Access control restricts template and data source access to authorized agents
- Audit logging tracks all document generation activities for compliance
- Sensitive data handling follows privacy and data protection regulations

## Configuration

The document-generator skill can be configured with the following parameters:

- `default_format`: Default output format (pdf, docx, html, markdown, text)
- `template_directory`: Directory for storing document templates (default: ./templates)
- `max_template_size`: Maximum template size (default: 1MB)
- `allowed_data_sources`: Allowed data source types (json, csv, database, api)
- `security_mode`: Template security mode (strict, moderate, relaxed)

This skill is essential for any agent that needs to generate professional documents, reports, contracts, invoices, or other structured content automatically from templates and data sources.