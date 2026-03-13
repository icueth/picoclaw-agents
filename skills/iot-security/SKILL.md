---
name: iot-security
description: Comprehensive IoT and smart home security monitoring and protection system for AI agents with threat detection and automated response capabilities
---

# IoT Security

This built-in skill provides comprehensive IoT and smart home security monitoring and protection capabilities for AI agents to detect, prevent, and respond to security threats across connected devices and networks.

## Capabilities

- **Device Security Assessment**: Assess security posture of IoT devices, firmware, and configurations
- **Network Monitoring**: Monitor network traffic for suspicious patterns, unauthorized devices, and data exfiltration
- **Threat Detection**: Detect known IoT threats, malware, botnets, and attack patterns using threat intelligence
- **Vulnerability Management**: Identify and manage vulnerabilities in IoT devices, firmware, and protocols
- **Access Control**: Enforce proper access controls, authentication, and authorization for IoT devices
- **Firmware Analysis**: Analyze device firmware for security vulnerabilities and backdoors
- **Protocol Security**: Assess security of IoT communication protocols (MQTT, CoAP, Zigbee, Z-Wave, Bluetooth)
- **Automated Response**: Implement automated responses to security incidents with containment and remediation
- **Compliance Checking**: Verify compliance with IoT security standards (NIST, ETSI, OWASP IoT)
- **Security Hardening**: Apply security hardening recommendations to improve device and network security

## Usage Examples

### IoT Device Security Assessment
```yaml
tool: iot-security
action: assess_devices
devices:
  - name: "Smart Camera"
    ip: "192.168.1.100"
    protocol: "http"
    firmware_version: "2.1.5"
  - name: "Smart Thermostat"
    ip: "192.168.1.101"
    protocol: "mqtt"
    firmware_version: "3.0.2"
  - name: "Smart Light Bulb"
    ip: "192.168.1.102"
    protocol: "zigbee"
    firmware_version: "1.5.0"
assessment_types:
  - "default_credentials"
  - "firmware_vulnerabilities"
  - "insecure_protocols"
  - "excessive_permissions"
  - "data_privacy_issues"
severity_threshold: "medium"
```

### Network Traffic Monitoring
```yaml
tool: iot-security
action: monitor_network
network_range: "192.168.1.0/24"
monitoring_types:
  - "unauthorized_devices"
  - "suspicious_traffic_patterns"
  - "data_exfiltration"
  - "command_and_control_communication"
  - "port_scanning"
  - "ddos_activity"
threat_intelligence_feeds:
  - "iot_malware_signatures"
  - "botnet_c2_servers"
  - "known_vulnerable_devices"
alert_threshold: "high"
response_actions:
  - condition: "threat_severity >= 'high'"
    action: "isolate_device"
  - condition: "data_exfiltration_detected = true"
    action: "block_traffic"
```

### Vulnerability Management
```yaml
tool: iot-security
action: manage_vulnerabilities
devices:
  - name: "Smart Camera"
    cve_ids: ["CVE-2023-1234", "CVE-2023-5678"]
    risk_level: "critical"
  - name: "Smart Thermostat"
    cve_ids: ["CVE-2023-9012"]
    risk_level: "medium"
remediation_strategies:
  - condition: "risk_level = 'critical'"
    actions:
      - "immediate_firmware_update"
      - "network_isolation"
      - "notify_administrator"
  - condition: "risk_level = 'medium'"
    actions:
      - "schedule_firmware_update"
      - "apply_network_rules"
      - "increase_monitoring"
patch_management: true
```

### Security Hardening
```yaml
tool: iot-security
action: harden_security
hardening_areas:
  - "network_segmentation"
  - "strong_authentication"
  - "encrypted_communications"
  - "regular_updates"
  - "minimal_permissions"
  - "secure_defaults"
recommendations:
  - type: "network"
    action: "create_iot_vlan"
    description: "Isolate IoT devices on separate VLAN"
  - type: "authentication"
    action: "enforce_strong_passwords"
    description: "Require strong passwords for all devices"
  - type: "updates"
    action: "enable_auto_updates"
    description: "Enable automatic security updates where available"
  - type: "monitoring"
    action: "deploy_network_monitoring"
    description: "Deploy continuous network monitoring for IoT traffic"
implementation_priority: "high"
```

## Security Considerations

- Security assessments run with minimal required permissions to prevent disruption
- Sensitive security findings are encrypted and access-controlled
- Automated responses are validated with safety checks before execution
- Access control ensures only authorized agents can perform security operations
- Audit logging tracks all security activities for compliance and forensic analysis
- Threat intelligence feeds are regularly updated from trusted sources

## Configuration

The iot-security skill can be configured with the following parameters:

- `default_assessment_depth`: Default assessment depth (basic, standard, comprehensive)
- `threat_intelligence_sources`: Enabled threat intelligence sources
- `auto_response_level`: Level of automated response (none, low_risk, medium_risk, all)
- `compliance_standards`: Enabled compliance standards (nist, etsi, owasp_iot)
- `monitoring_frequency`: Frequency of security monitoring (continuous, hourly, daily)

This skill is essential for any agent that needs to secure IoT devices, monitor network traffic, detect and respond to threats, or ensure compliance with IoT security standards in smart home and industrial environments.