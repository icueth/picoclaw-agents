---
name: financial-compliance
description: Comprehensive financial compliance and regulatory management system for AI agents with automated reporting, audit trails, and regulatory monitoring capabilities
---

# Financial Compliance

This built-in skill provides comprehensive financial compliance and regulatory management capabilities for AI agents to ensure adherence to financial regulations, maintain audit trails, generate compliance reports, and monitor regulatory changes across multiple jurisdictions.

## Capabilities

- **Regulatory Framework Support**: Support multiple regulatory frameworks (KYC/AML, GDPR, PCI-DSS, SOX, MiFID II, Dodd-Frank, Basel III)
- **Automated Reporting**: Generate automated compliance reports for regulatory submissions and internal audits
- **Transaction Monitoring**: Monitor financial transactions for suspicious activities and regulatory violations
- **Customer Due Diligence**: Implement customer due diligence (CDD) and enhanced due diligence (EDD) processes
- **Audit Trail Management**: Maintain comprehensive audit trails for all financial activities and compliance actions
- **Risk Assessment**: Conduct automated risk assessments for customers, transactions, and business activities
- **Regulatory Change Monitoring**: Monitor regulatory changes and automatically update compliance procedures
- **Data Privacy Management**: Ensure data privacy compliance with encryption, access controls, and data retention policies
- **Compliance Training**: Provide automated compliance training and certification tracking for employees
- **Integration Hub**: Integrate with compliance tools, regulatory databases, and external verification services

## Usage Examples

### KYC/AML Compliance
```yaml
tool: financial-compliance
action: perform_kyc_aml
customer:
  name: "John Doe"
  email: "john.doe@example.com"
  date_of_birth: "1985-06-15"
  address:
    street: "123 Main St"
    city: "San Francisco"
    state: "CA"
    postal_code: "94105"
    country: "US"
  identification:
    type: "passport"
    number: "P12345678"
    country: "US"
    expiry_date: "2030-12-31"
verification_checks:
  - type: "identity_verification"
    provider: "jumio"
    threshold: "high"
  - type: "sanctions_screening"
    provider: "world_check"
    lists: ["ofac", "un", "eu"]
  - type: "pep_screening"
    provider: "world_check"
    threshold: "medium"
  - type: "adverse_media"
    provider: "world_check"
    sources: ["news", "legal", "regulatory"]
risk_assessment:
  factors:
    - "geographic_risk"
    - "product_risk"
    - "transaction_risk"
    - "customer_profile_risk"
  overall_risk: "low"
```

### Regulatory Reporting
```yaml
tool: financial-compliance
action: generate_regulatory_report
report_type: "suspicious_activity_report"
jurisdiction: "us"
reporting_period:
  start: "2026-01-01"
  end: "2026-03-31"
data_sources:
  - "transaction_monitoring"
  - "customer_due_diligence"
  - "risk_assessment"
  - "audit_trails"
required_fields:
  - "customer_information"
  - "transaction_details"
  - "suspicious_activity_indicators"
  - "supporting_documentation"
  - "filing_entity_information"
submission_method: "automated_filing"
certification_required: true
```

### Transaction Monitoring
```yaml
tool: financial-compliance
action: monitor_transactions
monitoring_rules:
  - name: "large_cash_transaction"
    condition: "amount >= 10000 AND currency = 'USD'"
    action: "flag_for_review"
    severity: "high"
  - name: "structuring_attempt"
    condition: "multiple_transactions < 10000 AND total_amount >= 10000 AND same_customer AND within_24h"
    action: "flag_for_review"
    severity: "high"
  - name: "high_risk_jurisdiction"
    condition: "destination_country IN ['high_risk_list'] AND amount > 1000"
    action: "enhanced_due_diligence"
    severity: "medium"
  - name: "unusual_pattern"
    condition: "transaction_pattern_deviation > 3_sigma"
    action: "monitor_closely"
    severity: "low"
real_time_alerts: true
automated_investigation: true
```

### Audit Trail Management
```yaml
tool: financial-compliance
action: manage_audit_trail
activities:
  - type: "customer_onboarding"
    timestamp: "2026-03-13T14:30:00Z"
    user: "agent_compliance_officer"
    details: "Completed KYC verification for customer John Doe"
    evidence: ["/evidence/kyc_verification_12345.pdf"]
  - type: "transaction_approval"
    timestamp: "2026-03-13T15:45:00Z"
    user: "agent_risk_manager"
    details: "Approved high-value transaction after enhanced due diligence"
    evidence: ["/evidence/edd_report_67890.pdf"]
  - type: "compliance_report_filing"
    timestamp: "2026-03-13T16:30:00Z"
    user: "agent_compliance_system"
    details: "Filed monthly SAR report with FinCEN"
    evidence: ["/reports/sar_march_2026.xml"]
retention_policy: "7_years"
encryption_enabled: true
access_logging: true
```

## Security Considerations

- Compliance data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific compliance information
- Privacy controls allow users to manage what compliance data is shared and with whom
- Audit logging tracks all compliance activities for accountability and security
- Sensitive customer information is protected with appropriate access controls and data minimization principles
- Regulatory submissions are verified and certified before transmission

## Configuration

The financial-compliance skill can be configured with the following parameters:

- `default_jurisdictions`: Default regulatory jurisdictions (us, eu, global)
- `compliance_frameworks`: Enabled compliance frameworks (kyc_aml, gdpr, pci_dss, sox)
- `monitoring_frequency`: Frequency of compliance monitoring (real_time, hourly, daily)
- `reporting_schedule`: Automated reporting schedule (monthly, quarterly, annually)
- `privacy_level`: Privacy level for compliance data handling (strict, moderate, relaxed)

This skill is essential for any agent that needs to ensure financial regulatory compliance, monitor transactions for suspicious activities, generate compliance reports, maintain audit trails, or provide comprehensive financial compliance and regulatory management capabilities.