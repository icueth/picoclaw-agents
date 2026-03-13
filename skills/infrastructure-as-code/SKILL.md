---
name: infrastructure-as-code
description: Comprehensive infrastructure as code management system for AI agents with multi-platform support and automated deployment capabilities
---

# Infrastructure as Code

This built-in skill provides comprehensive infrastructure as code (IaC) management capabilities for AI agents to define, deploy, and manage cloud infrastructure using declarative configuration files across multiple platforms.

## Capabilities

- **Multi-Platform Support**: Support for Terraform, AWS CloudFormation, Azure ARM, Google Deployment Manager, Pulumi, and Crossplane
- **Infrastructure Definition**: Create and manage infrastructure definitions using YAML, JSON, HCL, or programming languages
- **State Management**: Manage infrastructure state with version control, locking, and drift detection
- **Automated Deployment**: Deploy infrastructure changes with automated validation, testing, and rollback capabilities
- **Security Scanning**: Scan infrastructure code for security vulnerabilities, misconfigurations, and compliance violations
- **Cost Optimization**: Analyze and optimize infrastructure costs with recommendations and budget alerts
- **Compliance Checking**: Verify infrastructure compliance with organizational policies and regulatory requirements
- **Template Management**: Create and manage reusable infrastructure templates and modules
- **Environment Management**: Manage multiple environments (dev, staging, prod) with consistent configurations
- **Integration with CI/CD**: Integrate infrastructure deployment into CI/CD pipelines with automated testing

## Usage Examples

### Terraform Infrastructure Deployment
```yaml
tool: infrastructure-as-code
action: deploy_infrastructure
platform: "terraform"
configuration:
  source: "/infra/terraform"
  variables:
    region: "us-west-2"
    environment: "production"
    instance_count: 3
  backend:
    type: "s3"
    bucket: "my-terraform-state"
    key: "production/terraform.tfstate"
validation: true
auto_approve: false
```

### CloudFormation Stack Management
```yaml
tool: infrastructure-as-code
action: manage_stack
platform: "cloudformation"
stack_name: "web-application-prod"
template_path: "/infra/cloudformation/web-app.yaml"
parameters:
  VpcId: "vpc-12345"
  InstanceType: "t3.medium"
  MinCapacity: 2
  MaxCapacity: 5
capabilities: ["CAPABILITY_IAM"]
tags:
  Environment: "production"
  Owner: "devops-team"
```

### Security and Compliance Scanning
```yaml
tool: infrastructure-as-code
action: scan_infrastructure
platform: "terraform"
source_path: "/infra/terraform"
scan_types:
  - "security_vulnerabilities"
  - "compliance_violations"
  - "cost_optimization"
  - "best_practices"
compliance_standards:
  - "cis_aws"
  - "pci_dss"
  - "hipaa"
severity_threshold: "medium"
```

### Multi-Environment Deployment
```yaml
tool: infrastructure-as-code
action: deploy_multi_environment
platform: "pulumi"
environments:
  - name: "development"
    config_file: "/infra/pulumi/dev.yaml"
    stack_name: "myapp-dev"
  - name: "staging"
    config_file: "/infra/pulumi/staging.yaml"
    stack_name: "myapp-staging"
  - name: "production"
    config_file: "/infra/pulumi/prod.yaml"
    stack_name: "myapp-prod"
sequential_deployment: true
approval_required: ["production"]
```

## Security Considerations

- Infrastructure code is scanned for security vulnerabilities before deployment
- State files are encrypted and access-controlled to prevent unauthorized modifications
- Deployment operations require appropriate permissions and approvals
- Audit logging tracks all infrastructure changes for compliance and security monitoring
- Sensitive configuration data is managed through secure secret management systems

## Configuration

The infrastructure-as-code skill can be configured with the following parameters:

- `default_platform`: Default IaC platform (terraform, cloudformation, pulumi, crossplane)
- `state_backend`: Default state backend configuration (s3, azurerm, gcs, local)
- `validation_enabled`: Enable automatic validation before deployment (default: true)
- `compliance_standards`: Enabled compliance standards for scanning
- `auto_approve_threshold`: Automatic approval threshold for non-production environments

This skill is essential for any agent that needs to manage cloud infrastructure, automate deployments, ensure security and compliance, or integrate infrastructure management into development workflows.