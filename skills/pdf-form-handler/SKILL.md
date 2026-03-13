---
name: pdf-form-handler
description: Intelligent PDF form processing and management system for AI agents with validation and automation capabilities
---

# PDF Form Handler

This built-in skill provides intelligent PDF form processing and management capabilities for AI agents to fill, extract, validate, and automate workflows with interactive PDF forms.

## Capabilities

- **Form Field Detection**: Automatically detect and identify all form fields in PDF documents
- **Form Filling**: Populate PDF forms with data from various sources (JSON, CSV, databases, APIs)
- **Data Extraction**: Extract filled form data and convert to structured formats (JSON, CSV, XML)
- **Form Validation**: Validate form data against field constraints, required fields, and business rules
- **Digital Signatures**: Apply and verify digital signatures on PDF forms
- **Form Flattening**: Convert interactive forms to static PDFs after completion
- **Template Management**: Create and manage PDF form templates with predefined fields and logic
- **Conditional Logic**: Handle conditional form fields and dynamic content based on user input
- **Batch Processing**: Process multiple forms simultaneously with consistent data mapping
- **Error Handling**: Provide detailed error reports for invalid or incomplete form data

## Usage Examples

### Fill PDF Form
```yaml
tool: pdf-form-handler
action: fill_form
pdf_path: "/forms/employment_application.pdf"
data:
  first_name: "John"
  last_name: "Doe"
  email: "john.doe@example.com"
  phone: "+1-555-123-4567"
  position: "Software Engineer"
  start_date: "2026-04-01"
  salary_expectation: "$85,000"
output_path: "/filled_forms/john_doe_application.pdf"
flatten: true
```

### Extract Form Data
```yaml
tool: pdf-form-handler
action: extract_data
pdf_path: "/submitted_forms/tax_return_2025.pdf"
output_format: "json"
include_metadata: true
validate_fields: true
```

### Validate Form Data
```yaml
tool: pdf-form-handler
action: validate_form
pdf_path: "/forms/contract_template.pdf"
test_data:
  client_name: "ABC Corp"
  contract_value: "$100,000"
  start_date: "2026-03-15"
  end_date: "2026-12-31"
validation_rules:
  - field: "contract_value"
    rule: "greater_than_0"
  - field: "start_date"
    rule: "before_end_date"
  - field: "client_name"
    rule: "required_not_empty"
```

### Batch Form Processing
```yaml
tool: pdf-form-handler
action: batch_process
template_path: "/forms/invoice_template.pdf"
data_source:
  type: "csv"
  path: "/data/invoice_data.csv"
output_directory: "/invoices/march_2026/"
naming_pattern: "invoice_{{customer_id}}.pdf"
flatten: true
```

## Security Considerations

- Form data is encrypted at rest and in transit using industry-standard encryption
- Sensitive form fields (SSN, credit card numbers) are automatically detected and protected
- Access control ensures only authorized agents can process specific form types
- Audit logging tracks all form operations for compliance and security monitoring
- Input sanitization prevents injection attacks and malicious content in form fields

## Configuration

The pdf-form-handler skill can be configured with the following parameters:

- `default_output_format`: Default output format for extracted data (json, csv, xml)
- `auto_flatten`: Automatically flatten forms after filling (default: false)
- `max_form_size`: Maximum form size for processing (default: 25MB)
- `sensitive_field_detection`: Enable automatic detection of sensitive fields (default: true)
- `validation_level`: Form validation level (strict, moderate, relaxed)

This skill is essential for any agent that needs to handle PDF forms, automate document workflows, extract structured data from filled forms, or ensure compliance with form validation requirements.