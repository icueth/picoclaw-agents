---
name: document-security
description: Comprehensive document security and protection system for AI agents with encryption, redaction, and access control
---

# Document Security

This built-in skill provides comprehensive document security and protection capabilities for AI agents to secure sensitive documents, redact confidential information, and manage access controls.

## Capabilities

- **Document Encryption**: Encrypt documents with password protection and digital certificates
- **Redaction Tools**: Automatically redact sensitive information (PII, financial data, classified content)
- **Watermarking**: Add visible and invisible watermarks for document tracking and authentication
- **Access Control**: Implement role-based access control and permission management for documents
- **Digital Signatures**: Apply and verify digital signatures for document authenticity and integrity
- **Metadata Sanitization**: Remove or sanitize metadata that may contain sensitive information
- **Audit Logging**: Track all document access, modifications, and security operations
- **Compliance Checking**: Verify document security against regulatory requirements (GDPR, HIPAA, etc.)
- **Secure Sharing**: Generate secure sharing links with expiration dates and access restrictions
- **Threat Detection**: Detect potentially malicious or compromised documents

## Usage Examples

### Encrypt Document
```yaml
tool: document-security
action: encrypt_document
document_path: "/reports/financial_report_q1.pdf"
encryption_method: "aes-256"
password: "{{secure_password}}"
output_path: "/secure/financial_report_q1_encrypted.pdf"
preserve_metadata: false
```

### Redact Sensitive Information
```yaml
tool: document-security
action: redact_document
document_path: "/documents/employee_record.pdf"
redaction_rules:
  - pattern: "\\d{3}-\\d{2}-\\d{4}"
    label: "SSN"
    replacement: "[REDACTED]"
  - pattern: "\\d{4} \\d{4} \\d{4} \\d{4}"
    label: "Credit Card"
    replacement: "[REDACTED]"
  - keywords: ["confidential", "secret", "classified"]
    label: "Classification"
    replacement: "[CLASSIFIED]"
output_path: "/redacted/employee_record_redacted.pdf"
```

### Apply Digital Signature
```yaml
tool: document-security
action: sign_document
document_path: "/contracts/agreement.pdf"
certificate_path: "/certs/company_cert.p12"
certificate_password: "{{cert_password}}"
signature_reason: "Contract approval"
signature_location: "Page 1, bottom right"
output_path: "/signed/agreement_signed.pdf"
```

### Sanitize Metadata
```yaml
tool: document-security
action: sanitize_metadata
document_path: "/presentations/internal_deck.pptx"
remove_fields:
  - "author"
  - "created_by"
  - "last_modified_by"
  - "comments"
  - "track_changes"
output_path: "/clean/internal_deck_sanitized.pptx"
```

## Security Considerations

- All encryption operations use industry-standard algorithms (AES-256, RSA-2048)
- Password handling follows secure credential management practices
- Redaction patterns are validated to prevent false positives/negatives
- Access control integrates with existing authentication systems
- Audit logs are tamper-proof and comply with regulatory requirements
- Secure key management prevents unauthorized access to encryption keys

## Configuration

The document-security skill can be configured with the following parameters:

- `default_encryption_method`: Default encryption method (aes-256, aes-128, des)
- `redaction_confidence_threshold`: Confidence threshold for automatic redaction (default: 0.9)
- `compliance_standards`: Enabled compliance standards (gdpr, hipaa, pci-dss, soc2)
- `audit_log_retention`: Audit log retention period (default: 365 days)
- `key_management_system`: Key management system integration (local, vault, cloud)

This skill is essential for any agent that needs to handle sensitive documents, ensure regulatory compliance, protect confidential information, or implement enterprise-grade document security workflows.