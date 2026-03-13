---
name: payment-processing
description: Secure payment processing and transaction management system for AI agents with multi-gateway support and fraud detection capabilities
---

# Payment Processing

This built-in skill provides secure payment processing and transaction management capabilities for AI agents to handle payments, refunds, subscriptions, and financial transactions across multiple payment gateways and methods.

## Capabilities

- **Multi-Gateway Support**: Integrate with popular payment gateways (Stripe, PayPal, Square, Braintree, Adyen, Authorize.net)
- **Payment Method Support**: Support multiple payment methods (credit/debit cards, digital wallets, bank transfers, cryptocurrency)
- **Fraud Detection**: Implement advanced fraud detection and prevention with machine learning and rule-based systems
- **Subscription Management**: Manage recurring payments and subscription billing with automated renewal and dunning
- **Refund Processing**: Handle refund requests with automated approval workflows and payment gateway integration
- **Compliance and Security**: Ensure PCI-DSS compliance and implement strong security measures for payment data
- **Transaction Monitoring**: Monitor transactions in real-time with alerts and anomaly detection
- **Currency and Exchange**: Handle multiple currencies with real-time exchange rates and conversion
- **Reporting and Analytics**: Provide payment analytics with transaction trends, success rates, and revenue metrics
- **Dispute Management**: Handle chargebacks and disputes with evidence collection and response automation

## Usage Examples

### Payment Processing
```yaml
tool: payment-processing
action: process_payment
gateway: "stripe"
payment_method:
  type: "credit_card"
  card_number: "{{encrypted_card_number}}"
  expiry_month: "12"
  expiry_year: "2026"
  cvv: "{{encrypted_cvv}"
amount: 99.99
currency: "USD"
customer:
  email: "customer@example.com"
  name: "John Doe"
  address:
    line1: "123 Main St"
    city: "San Francisco"
    state: "CA"
    postal_code: "94105"
    country: "US"
fraud_check: true
metadata:
  order_id: "ORD-12345"
  product_id: "PROD-67890"
  source: "web_store"
```

### Subscription Management
```yaml
tool: payment-processing
action: manage_subscription
gateway: "stripe"
subscription:
  customer_id: "cus_12345"
  plan_id: "premium_monthly"
  quantity: 1
  trial_period_days: 14
  billing_cycle_anchor: "now"
  payment_behavior: "default_incomplete"
  metadata:
    user_id: "user_67890"
    account_type: "business"
automation_rules:
  - event: "payment_failed"
    action: "retry_payment"
    attempts: 3
    delay: "3d"
  - event: "trial_ending"
    action: "send_trial_reminder"
    channels: ["email", "sms"]
  - event: "subscription_cancelled"
    action: "deactivate_account"
    grace_period: "7d"
```

### Fraud Detection and Prevention
```yaml
tool: payment-processing
action: detect_fraud
transaction:
  amount: 499.99
  currency: "USD"
  customer:
    email: "new_customer@example.com"
    ip_address: "203.0.113.42"
    device_fingerprint: "df_12345"
    location: "New York, US"
  payment_method:
    type: "credit_card"
    last4: "4242"
    country: "US"
    issuer: "Chase"
risk_factors:
  - "high_amount_for_new_customer"
  - "different_billing_shipping_location"
  - "multiple_failed_attempts_recent"
  - "velocity_check_failed"
fraud_score: 0.75
recommendation: "manual_review"
```

### Refund Processing
```yaml
tool: payment-processing
action: process_refund
gateway: "paypal"
refund:
  transaction_id: "txn_12345"
  amount: 99.99
  currency: "USD"
  reason: "customer_request"
  customer_note: "Item not as described"
  merchant_note: "Customer requested refund due to product description mismatch"
  metadata:
    order_id: "ORD-12345"
    return_tracking: "UPS-123456789"
approval_workflow:
  - condition: "amount <= 100"
    action: "auto_approve"
  - condition: "amount > 100"
    action: "manual_approval_required"
```

## Security Considerations

- Payment data is never stored locally and is handled through PCI-DSS compliant payment gateways
- Sensitive payment information is encrypted using industry-standard encryption before transmission
- Access control ensures only authorized agents can process payments or access transaction data
- Fraud detection systems prevent unauthorized or suspicious transactions
- Audit logging tracks all payment processing activities for compliance and security monitoring
- Integration credentials are securely stored and never exposed in logs

## Configuration

The payment-processing skill can be configured with the following parameters:

- `default_gateway`: Default payment gateway (stripe, paypal, square, braintree)
- `fraud_detection_level`: Level of fraud detection (basic, standard, comprehensive)
- `currency_support`: Enabled currencies for transactions
- `compliance_regions`: Enabled compliance regions (us, eu, global)
- `security_level`: Security level for payment processing (standard, high, enterprise)

This skill is essential for any agent that needs to process payments securely, manage subscriptions, handle refunds, detect fraud, or provide comprehensive payment processing and transaction management capabilities.