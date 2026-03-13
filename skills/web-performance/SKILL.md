---
name: web-performance
description: Advanced web performance optimization and monitoring system for AI agents with comprehensive analysis and automated improvement capabilities
---

# Web Performance

This built-in skill provides advanced web performance optimization and monitoring capabilities for AI agents to analyze, optimize, and monitor web application performance across multiple dimensions and user experiences.

## Capabilities

- **Core Web Vitals**: Measure and optimize Core Web Vitals (LCP, FID, CLS, INP) with detailed analysis and recommendations
- **Performance Budgeting**: Define and enforce performance budgets for page weight, load time, and resource usage
- **Bundle Analysis**: Analyze JavaScript and CSS bundle sizes, dependencies, and optimization opportunities
- **Image Optimization**: Optimize images with modern formats (WebP, AVIF), responsive sizing, and lazy loading
- **Caching Strategies**: Implement intelligent caching strategies (CDN, browser, service worker) for optimal performance
- **Critical Rendering Path**: Optimize critical rendering path with proper resource loading and execution order
- **Network Optimization**: Optimize network requests with compression, HTTP/2, and resource prioritization
- **Performance Monitoring**: Monitor real-user performance metrics with alerts and trend analysis
- **Synthetic Testing**: Perform synthetic performance testing across multiple locations and devices
- **Automated Optimization**: Apply automated performance optimizations with safe rollback mechanisms

## Usage Examples

### Core Web Vitals Analysis
```yaml
tool: web-performance
action: analyze_core_web_vitals
url: "https://my-app.com"
metrics:
  - "lcp"
  - "fid"
  - "cls"
  - "inp"
  - "ttfb"
  - "fcp"
device_types:
  - "mobile"
  - "desktop"
locations:
  - "us-west"
  - "eu-central"
  - "ap-southeast"
thresholds:
  lcp: "2.5s"
  fid: "100ms"
  cls: "0.1"
  inp: "200ms"
```

### Bundle Analysis and Optimization
```yaml
tool: web-performance
action: analyze_bundle
bundle_path: "/dist/main.js"
analysis_types:
  - "size_breakdown"
  - "duplicate_dependencies"
  - "unused_code"
  - "large_dependencies"
optimization_strategies:
  - "code_splitting"
  - "tree_shaking"
  - "lazy_loading"
  - "compression"
  - "cdn_integration"
output_format: "interactive_html"
```

### Image Optimization
```yaml
tool: web-performance
action: optimize_images
source_directory: "/src/images"
target_directory: "/dist/images"
optimizations:
  - format: "webp"
    quality: 85
  - format: "avif"
    quality: 75
  - responsive: true
    sizes: ["320w", "640w", "1024w", "1920w"]
  - lazy_loading: true
  - progressive_loading: true
```

### Performance Budget Enforcement
```yaml
tool: web-performance
action: enforce_performance_budget
budget_file: "/performance-budget.json"
budget_rules:
  - resource_type: "javascript"
    max_size: "500kb"
    max_count: 10
  - resource_type: "css"
    max_size: "100kb"
    max_count: 5
  - resource_type: "image"
    max_size: "200kb"
    max_count: 20
  - metric: "total_page_weight"
    max_size: "2mb"
  - metric: "load_time"
    max_time: "3s"
ci_integration: true
```

## Security Considerations

- Performance analysis runs in isolated environments to prevent data leakage
- Sensitive performance data is encrypted and access-controlled
- Automated optimizations are validated with safety checks before application
- Access control ensures only authorized agents can modify performance configurations
- Audit logging tracks all performance optimization activities for compliance

## Configuration

The web-performance skill can be configured with the following parameters:

- `default_metrics`: Default performance metrics to track (core_web_vitals, custom_metrics)
- `optimization_level`: Level of automated optimization (conservative, moderate, aggressive)
- `monitoring_frequency`: Frequency of performance monitoring (real_time, hourly, daily)
- `alert_thresholds`: Alert thresholds for performance degradation
- `rollback_enabled`: Enable automatic rollback for failed optimizations (default: true)

This skill is essential for any agent that needs to optimize web application performance, ensure fast user experiences, monitor performance metrics, or implement performance best practices in development workflows.