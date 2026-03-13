---
name: data-analyst
description: Comprehensive data analysis and visualization system for AI agents with support for multiple data formats and statistical methods
---

# Data Analyst

This built-in skill provides comprehensive data analysis and visualization capabilities for AI agents to process, analyze, and visualize data from various sources using advanced statistical methods and machine learning techniques.

## Capabilities

- **Data Import**: Import data from various formats (CSV, JSON, Excel, SQL databases, APIs)
- **Data Cleaning**: Clean and preprocess data including handling missing values, outliers, and duplicates
- **Statistical Analysis**: Perform descriptive statistics, hypothesis testing, and correlation analysis
- **Data Visualization**: Create interactive charts, graphs, and dashboards using various visualization libraries
- **Machine Learning**: Apply machine learning algorithms for classification, regression, and clustering
- **Time Series Analysis**: Analyze time series data with forecasting and trend detection
- **Report Generation**: Generate comprehensive analysis reports with insights and recommendations
- **Data Export**: Export results to various formats (CSV, JSON, PDF, HTML, images)
- **Real-time Analysis**: Perform real-time data analysis on streaming data sources
- **Database Integration**: Connect to and query various database systems (MySQL, PostgreSQL, SQLite, MongoDB)

## Usage Examples

### Basic Data Analysis
```yaml
tool: data-analyst
action: analyze_data
data_source:
  type: "csv"
  path: "/data/sales.csv"
analysis_type: "descriptive"
include:
  - "summary_statistics"
  - "correlation_matrix"
  - "distribution_plots"
```

### Machine Learning
```yaml
tool: data-analyst
action: apply_ml
data_source:
  type: "database"
  connection: "postgresql://user:pass@localhost/analytics"
  query: "SELECT * FROM customer_data WHERE created_at > '2025-01-01'"
ml_task: "classification"
target_column: "churn_risk"
features: ["age", "spend", "support_tickets", "login_frequency"]
algorithm: "random_forest"
```

### Time Series Forecasting
```yaml
tool: data-analyst
action: forecast_time_series
data_source:
  type: "api"
  url: "https://api.example.com/metrics/daily"
  headers:
    Authorization: "Bearer {{api_token}}"
forecast_horizon: 30
confidence_interval: 0.95
seasonality: "weekly"
```

## Security Considerations

- Data privacy protection with encryption for sensitive datasets
- Access control ensures only authorized agents can access specific data sources
- Secure credential management for database connections and API authentication
- Audit logging tracks all data analysis operations for compliance
- Data anonymization options for handling personally identifiable information (PII)

## Configuration

The data-analyst skill can be configured with the following parameters:

- `default_visualization_library`: Default visualization library (matplotlib, plotly, seaborn)
- `max_data_size`: Maximum data size to process (default: 1GB)
- `cache_results`: Enable result caching (default: true)
- `privacy_mode`: Privacy mode for handling sensitive data (strict, moderate, relaxed)
- `compute_backend`: Compute backend (local, distributed, cloud)
- `export_formats`: Supported export formats (csv, json, pdf, html, png, svg)

This skill is essential for any agent that needs to analyze data, generate insights, create visualizations, or apply machine learning techniques. It provides comprehensive data analysis capabilities while maintaining data security and privacy.