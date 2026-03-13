---
name: trading-analytics
description: Advanced trading and investment analytics system for AI agents with market analysis, portfolio optimization, and risk management capabilities
---

# Trading Analytics

This built-in skill provides advanced trading and investment analytics capabilities for AI agents to analyze markets, optimize portfolios, manage risk, and provide intelligent trading insights and recommendations.

## Capabilities

- **Market Analysis**: Analyze financial markets with technical indicators, fundamental analysis, and sentiment analysis
- **Portfolio Optimization**: Optimize investment portfolios using modern portfolio theory, risk parity, and other optimization techniques
- **Risk Management**: Implement comprehensive risk management strategies with position sizing, stop losses, and diversification
- **Trading Strategy Development**: Develop and backtest trading strategies with historical data and statistical validation
- **Real-time Monitoring**: Monitor markets and portfolios in real-time with alerts and automated responses
- **Alternative Data Integration**: Integrate alternative data sources (social media, web traffic, satellite imagery) for alpha generation
- **Tax Optimization**: Optimize trading for tax efficiency with tax-loss harvesting and lot selection
- **Performance Attribution**: Analyze trading performance with attribution models and benchmark comparisons
- **Regulatory Compliance**: Ensure trading activities comply with relevant regulations and reporting requirements
- **Machine Learning Models**: Apply machine learning models for price prediction, pattern recognition, and anomaly detection

## Usage Examples

### Market Analysis and Signals
```yaml
tool: trading-analytics
action: analyze_market
assets:
  - symbol: "AAPL"
    timeframe: "daily"
    indicators:
      - "moving_averages"
      - "rsi"
      - "macd"
      - "bollinger_bands"
      - "volume_analysis"
  - symbol: "SPY"
    timeframe: "weekly"
    indicators:
      - "sector_rotation"
      - "market_breadth"
      - "volatility_analysis"
sentiment_analysis:
  sources: ["twitter", "reddit", "news"]
  weight: 0.3
technical_weight: 0.5
fundamental_weight: 0.2
signals:
  - condition: "price > 200_ma AND rsi < 30"
    action: "buy"
    confidence: "high"
  - condition: "price < 50_ma AND macd_histogram < 0"
    action: "sell"
    confidence: "medium"
```

### Portfolio Optimization
```yaml
tool: trading-analytics
action: optimize_portfolio
current_portfolio:
  - asset: "AAPL"
    allocation: 25
  - asset: "MSFT"
    allocation: 20
  - asset: "GOOGL"
    allocation: 15
  - asset: "AMZN"
    allocation: 15
  - asset: "TSLA"
    allocation: 10
  - asset: "BTC"
    allocation: 15
optimization_objective: "maximize_sharpe_ratio"
constraints:
  - "max_single_position <= 30"
  - "sector_concentration <= 40"
  - "volatility_target <= 15"
  - "esg_score >= 70"
risk_models:
  - "historical_covariance"
  - "factor_model"
  - "black_litterman"
recommended_allocation:
  - asset: "AAPL"
    allocation: 22
  - asset: "MSFT"
    allocation: 18
  - asset: "GOOGL"
    allocation: 16
  - asset: "AMZN"
    allocation: 14
  - asset: "NVDA"
    allocation: 12
  - asset: "BTC"
    allocation: 10
  - asset: "BND"
    allocation: 8
```

### Risk Management
```yaml
tool: trading-analytics
action: manage_risk
portfolio:
  - asset: "AAPL"
    position_size: 100
    entry_price: 175.50
    current_price: 180.25
    volatility: 0.25
  - asset: "BTC"
    position_size: 0.5
    entry_price: 45000
    current_price: 47500
    volatility: 0.80
risk_parameters:
  - metric: "var_95"
    threshold: 0.05
  - metric: "max_drawdown"
    threshold: 0.15
  - metric: "sharpe_ratio"
    threshold: 1.0
risk_controls:
  - type: "position_sizing"
    method: "kelly_criterion"
    fraction: 0.5
  - type: "stop_loss"
    method: "atr_multiple"
    multiple: 2.0
  - type: "diversification"
    method: "correlation_limit"
    threshold: 0.7
```

### Trading Strategy Backtesting
```yaml
tool: trading-analytics
action: backtest_strategy
strategy:
  name: "Mean Reversion SPY"
  logic: |
    if RSI(14) < 30 and price > 200_SMA:
        buy()
    if RSI(14) > 70 or price < 50_SMA:
        sell()
parameters:
  - name: "rsi_period"
    value: 14
  - name: "sma_period"
    value: 200
  - name: "position_size"
    value: 0.1
backtest_period:
  start: "2016-01-01"
  end: "2026-03-13"
benchmark: "SPY"
metrics:
  - "total_return"
  - "annualized_return"
  - "max_drawdown"
  - "sharpe_ratio"
  - "win_rate"
  - "profit_factor"
results:
  total_return: 125.5
  annualized_return: 8.7
  max_drawdown: -15.2
  sharpe_ratio: 1.2
  win_rate: 0.65
  profit_factor: 1.8
```

## Security Considerations

- Trading data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific trading information
- Privacy controls allow users to manage what trading data is shared and with whom
- Audit logging tracks all trading analytics activities for accountability and security
- Integration credentials are securely stored and never exposed in logs
- Trading recommendations require explicit user confirmation before execution

## Configuration

The trading-analytics skill can be configured with the following parameters:

- `default_timeframes`: Default timeframes for analysis (daily, weekly, monthly)
- `risk_tolerance`: Default risk tolerance level (conservative, moderate, aggressive)
- `data_sources`: Enabled data sources (yahoo_finance, alpha_vantage, polygon, alternative_data)
- `compliance_regions`: Enabled compliance regions (us, eu, asia)
- `privacy_level`: Privacy level for trading data sharing (private, selective, public)

This skill is essential for any agent that needs to analyze financial markets, optimize investment portfolios, manage trading risk, develop trading strategies, or provide intelligent trading insights and recommendations.