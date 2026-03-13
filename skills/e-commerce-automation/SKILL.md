---
name: e-commerce-automation
description: Comprehensive e-commerce automation and management system for AI agents with product management, order processing, and customer service capabilities
---

# E-commerce Automation

This built-in skill provides comprehensive e-commerce automation and management capabilities for AI agents to manage online stores, process orders, handle customer service, and optimize e-commerce operations across multiple platforms.

## Capabilities

- **Multi-Platform Support**: Integrate with popular e-commerce platforms (Shopify, WooCommerce, Magento, BigCommerce, Amazon, Etsy)
- **Product Management**: Manage product catalogs with automated inventory updates, pricing optimization, and content generation
- **Order Processing**: Automate order processing with payment verification, fulfillment coordination, and shipping integration
- **Customer Service**: Handle customer inquiries, returns, and complaints with intelligent responses and escalation workflows
- **Marketing Automation**: Automate marketing campaigns with email sequences, social media posts, and promotional campaigns
- **Analytics and Reporting**: Provide e-commerce analytics with sales trends, customer behavior, and performance metrics
- **Inventory Management**: Manage inventory levels with automated reordering, stock alerts, and supplier coordination
- **Payment Processing**: Handle payment processing with multiple payment methods, fraud detection, and refund management
- **SEO and Content Optimization**: Optimize product listings and content for search engines and conversion rates
- **Compliance and Tax Management**: Ensure compliance with e-commerce regulations, tax requirements, and data protection laws

## Usage Examples

### Product Management Automation
```yaml
tool: e-commerce-automation
action: manage_products
platform: "shopify"
products:
  - sku: "TSHIRT-BLUE-L"
    name: "Blue T-Shirt - Large"
    price: 29.99
    inventory: 50
    description: "Comfortable cotton t-shirt in blue"
    categories: ["clothing", "t-shirts"]
    images: ["/images/tshirt-blue-l.jpg"]
  - sku: "TSHIRT-RED-M"
    name: "Red T-Shirt - Medium"
    price: 29.99
    inventory: 25
    description: "Comfortable cotton t-shirt in red"
    categories: ["clothing", "t-shirts"]
    images: ["/images/tshirt-red-m.jpg"]
automation_rules:
  - condition: "inventory < 10"
    action: "send_low_stock_alert"
    recipients: ["inventory_manager"]
  - condition: "competitor_price < our_price * 0.95"
    action: "adjust_price"
    parameters:
      new_price: "competitor_price * 1.02"
  - condition: "season = 'summer'"
    action: "update_description"
    parameters:
      template: "Perfect for summer! {{product_name}} keeps you cool and comfortable."
```

### Order Processing Automation
```yaml
tool: e-commerce-automation
action: process_orders
platform: "woocommerce"
order_types:
  - type: "standard"
    payment_methods: ["credit_card", "paypal"]
    fulfillment: "standard_shipping"
    verification: "automated"
  - type: "express"
    payment_methods: ["credit_card"]
    fulfillment: "express_shipping"
    verification: "manual_review"
automation_workflows:
  - trigger: "new_order"
    conditions:
      - "payment_verified = true"
      - "fraud_score < 0.5"
    actions:
      - "create_fulfillment_request"
      - "send_order_confirmation"
      - "update_inventory"
  - trigger: "shipping_update"
    conditions:
      - "tracking_number != null"
    actions:
      - "send_shipping_notification"
      - "update_order_status"
```

### Customer Service Automation
```yaml
tool: e-commerce-automation
action: handle_customer_service
platform: "shopify"
service_types:
  - type: "order_inquiry"
    response_template: "order_status_template"
    escalation_threshold: "24h"
  - type: "return_request"
    response_template: "return_policy_template"
    approval_required: true
  - type: "product_question"
    response_template: "product_info_template"
    knowledge_base: "product_catalog"
automation_rules:
  - condition: "sentiment_score < 0.3"
    action: "escalate_to_human"
    priority: "high"
  - condition: "inquiry_type = 'return' AND order_age < 30"
    action: "auto_approve_return"
    notification: "send_return_label"
  - condition: "common_question = true"
    action: "send_automated_response"
    confidence_threshold: 0.8
```

### Marketing Automation
```yaml
tool: e-commerce-automation
action: automate_marketing
campaigns:
  - name: "Abandoned Cart Recovery"
    trigger: "cart_abandoned"
    delay: "1h"
    channels: ["email", "sms"]
    messages:
      - delay: "1h"
        template: "abandoned_cart_reminder"
      - delay: "24h"
        template: "abandoned_cart_discount"
        discount: "10%"
  - name: "Post-Purchase Follow-up"
    trigger: "order_delivered"
    delay: "3d"
    channels: ["email"]
    messages:
      - delay: "3d"
        template: "product_review_request"
      - delay: "7d"
        template: "cross_sell_recommendation"
analytics:
  - "conversion_rates"
  - "customer_lifetime_value"
  - "campaign_roi"
  - "email_open_rates"
```

## Security Considerations

- E-commerce data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific e-commerce information
- Privacy controls allow users to manage what customer data is shared and with whom
- Audit logging tracks all e-commerce automation activities for accountability and security
- Payment processing complies with PCI-DSS standards and never stores sensitive payment data
- Integration credentials are securely stored and never exposed in logs

## Configuration

The e-commerce-automation skill can be configured with the following parameters:

- `default_platforms`: Default e-commerce platforms (shopify, woocommerce, amazon, etsy)
- `automation_level`: Level of automation (basic, standard, comprehensive)
- `fraud_detection_enabled`: Enable fraud detection by default (default: true)
- `compliance_regions`: Enabled compliance regions (us, eu, asia, global)
- `privacy_level`: Privacy level for customer data handling (strict, moderate, relaxed)

This skill is essential for any agent that needs to manage e-commerce operations, automate order processing, handle customer service, optimize marketing campaigns, or provide comprehensive e-commerce automation and management capabilities.