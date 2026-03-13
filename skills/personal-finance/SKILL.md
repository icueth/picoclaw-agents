---
name: personal-finance
description: Comprehensive personal finance management and budgeting system for AI agents with expense tracking, investment monitoring, and financial planning capabilities
---

# Personal Finance

This built-in skill provides comprehensive personal finance management and budgeting capabilities for AI agents to track expenses, monitor investments, create budgets, and provide intelligent financial planning and advice.

## Capabilities

- **Expense Tracking**: Automatically track and categorize expenses from bank accounts, credit cards, and receipts
- **Budget Management**: Create and manage monthly budgets with category limits, alerts, and progress tracking
- **Investment Monitoring**: Monitor investment portfolios across stocks, bonds, mutual funds, and cryptocurrencies
- **Financial Goal Planning**: Set and track financial goals (emergency fund, retirement, major purchases)
- **Bill Management**: Track upcoming bills and payments with automatic reminders and payment scheduling
- **Credit Score Monitoring**: Monitor credit scores and provide recommendations for improvement
- **Tax Preparation**: Assist with tax preparation by organizing deductible expenses and generating reports
- **Financial Analytics**: Provide financial analytics and insights with trend analysis and spending patterns
- **Integration Hub**: Integrate with popular finance tools (Mint, YNAB, Personal Capital, brokerage accounts)
- **Security and Privacy**: Ensure security and privacy of financial data with encryption and access controls

## Usage Examples

### Budget Creation and Management
```yaml
tool: personal-finance
action: create_budget
budget_name: "Monthly Budget - March 2026"
income:
  salary: 5000
  freelance: 1000
  other: 200
categories:
  - name: "Housing"
    budgeted: 1500
    actual: 1450
    subcategories:
      - "Rent": 1200
      - "Utilities": 200
      - "Internet": 100
  - name: "Food"
    budgeted: 600
    actual: 650
    subcategories:
      - "Groceries": 400
      - "Dining Out": 250
  - name: "Transportation"
    budgeted: 300
    actual: 280
    subcategories:
      - "Gas": 150
      - "Public Transit": 50
      - "Car Maintenance": 100
  - name: "Entertainment"
    budgeted: 200
    actual: 180
    subcategories:
      - "Streaming Services": 50
      - "Events": 130
alerts:
  - category: "Food"
    threshold: 90
    notification: "email"
  - category: "Entertainment"
    threshold: 100
    notification: "push"
```

### Investment Portfolio Monitoring
```yaml
tool: personal-finance
action: monitor_investments
portfolio:
  - asset: "AAPL"
    type: "stock"
    shares: 100
    current_price: 175.50
    purchase_price: 150.25
    allocation: 25
  - asset: "VTI"
    type: "etf"
    shares: 50
    current_price: 225.75
    purchase_price: 200.00
    allocation: 30
  - asset: "BTC"
    type: "cryptocurrency"
    units: 0.5
    current_price: 45000
    purchase_price: 35000
    allocation: 15
  - asset: "BND"
    type: "bond_fund"
    shares: 100
    current_price: 75.25
    purchase_price: 78.50
    allocation: 30
performance_metrics:
  - "total_return"
  - "year_to_date_return"
  - "risk_adjusted_return"
  - "diversification_score"
alerts:
  - condition: "allocation > 35"
    action: "rebalance_recommendation"
  - condition: "single_day_loss > 5"
    action: "loss_alert"
```

### Financial Goal Planning
```yaml
tool: personal-finance
action: plan_financial_goals
goals:
  - name: "Emergency Fund"
    target_amount: 15000
    current_amount: 8000
    timeline: "2026-12-31"
    monthly_contribution: 583
    priority: "high"
  - name: "Down Payment - House"
    target_amount: 100000
    current_amount: 25000
    timeline: "2028-06-30"
    monthly_contribution: 3125
    priority: "medium"
  - name: "Retirement"
    target_amount: 1000000
    current_amount: 150000
    timeline: "2045-12-31"
    monthly_contribution: 1200
    priority: "medium"
progress_tracking: true
adjustment_recommendations: true
```

### Expense Analysis and Insights
```yaml
tool: personal-finance
action: analyze_expenses
time_period: "last_6_months"
metrics:
  - "spending_by_category"
  - "monthly_trends"
  - "subscription_analysis"
  - "seasonal_patterns"
  - "budget_adherence"
insights:
  - "highest_spending_categories"
  - "recurring_subscriptions"
  - "unnecessary_expenses"
  - "saving_opportunities"
  - "spending_correlation_with_income"
visualization: "interactive_dashboard"
```

## Security Considerations

- Financial data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific financial information
- Privacy controls allow users to manage what financial data is shared and with whom
- Audit logging tracks all financial management activities for accountability and security
- Integration credentials are securely stored and never exposed in logs
- Sensitive financial transactions require explicit user confirmation before execution

## Configuration

The personal-finance skill can be configured with the following parameters:

- `default_currency`: Default currency for financial calculations (default: USD)
- `budget_categories`: Default budget categories and allocations
- `alert_thresholds`: Default alert thresholds for budget overruns and investment changes
- `integration_platforms`: Enabled integration platforms (mint, ynab, personal_capital, brokerages)
- `privacy_level`: Privacy level for financial data sharing (private, selective, public)

This skill is essential for any agent that needs to help users manage personal finances, track expenses and investments, create budgets, plan for financial goals, or provide intelligent financial planning and advice.