---
name: sensor-integration
description: Advanced sensor data integration and processing system for AI agents with multi-sensor support and real-time analytics capabilities
---

# Sensor Integration

This built-in skill provides advanced sensor data integration and processing capabilities for AI agents to collect, analyze, and act on data from various sensors and IoT devices in real-time.

## Capabilities

- **Multi-Sensor Support**: Support for temperature, humidity, motion, light, sound, air quality, soil moisture, water level, pressure, and custom sensors
- **Protocol Integration**: Integrate with various sensor protocols (MQTT, HTTP, CoAP, Bluetooth LE, Zigbee, Modbus, CAN bus)
- **Real-Time Data Processing**: Process sensor data in real-time with streaming analytics and anomaly detection
- **Data Aggregation**: Aggregate data from multiple sensors and sources with time-series analysis
- **Alert Generation**: Generate alerts based on sensor thresholds, anomalies, and patterns
- **Predictive Analytics**: Apply machine learning models to predict trends, failures, and optimal conditions
- **Visualization**: Create real-time dashboards and visualizations for sensor data monitoring
- **Historical Analysis**: Store and analyze historical sensor data for trend identification and reporting
- **Edge Computing**: Process sensor data at the edge for low-latency responses and bandwidth optimization
- **Integration with Automation**: Trigger automation workflows based on sensor data and conditions

## Usage Examples

### Real-Time Environmental Monitoring
```yaml
tool: sensor-integration
action: monitor_environment
sensors:
  - type: "temperature"
    name: "Living Room Temperature"
    location: "living_room"
    protocol: "mqtt"
    topic: "sensors/living_room/temperature"
  - type: "humidity"
    name: "Living Room Humidity"
    location: "living_room"
    protocol: "mqtt"
    topic: "sensors/living_room/humidity"
  - type: "air_quality"
    name: "Living Room Air Quality"
    location: "living_room"
    protocol: "http"
    endpoint: "http://air-quality.local/api/data"
thresholds:
  temperature:
    min: 18
    max: 26
  humidity:
    min: 30
    max: 60
  air_quality:
    max: 50  # AQI
alert_channels: ["push", "email"]
```

### Predictive Maintenance
```yaml
tool: sensor-integration
action: predictive_maintenance
sensors:
  - type: "vibration"
    name: "Motor Vibration"
    equipment: "pump_001"
  - type: "temperature"
    name: "Motor Temperature"
    equipment: "pump_001"
  - type: "current"
    name: "Motor Current"
    equipment: "pump_001"
ml_model: "pump_failure_prediction_v2"
prediction_window: "7d"
confidence_threshold: 0.8
actions:
  - condition: "failure_probability > 0.7"
    action: "schedule_maintenance"
    priority: "high"
  - condition: "failure_probability > 0.4"
    action: "increase_monitoring"
    frequency: "5m"
```

### Smart Agriculture
```yaml
tool: sensor-integration
action: monitor_agriculture
sensors:
  - type: "soil_moisture"
    name: "Field A Soil Moisture"
    location: "field_a"
    depth: "10cm"
  - type: "temperature"
    name: "Field A Temperature"
    location: "field_a"
  - type: "light"
    name: "Field A Light Intensity"
    location: "field_a"
  - type: "rainfall"
    name: "Field A Rainfall"
    location: "field_a"
irrigation_rules:
  - condition: "soil_moisture < 30 AND forecast_rain = false"
    action: "start_irrigation"
    duration: "30m"
  - condition: "temperature > 35"
    action: "start_irrigation"
    duration: "15m"
```

### Industrial Monitoring
```yaml
tool: sensor-integration
action: monitor_industrial
sensors:
  - type: "pressure"
    name: "Boiler Pressure"
    system: "boiler_001"
  - type: "temperature"
    name: "Boiler Temperature"
    system: "boiler_001"
  - type: "flow_rate"
    name: "Water Flow Rate"
    system: "boiler_001"
  - type: "vibration"
    name: "Pump Vibration"
    system: "pump_001"
safety_thresholds:
  pressure:
    warning: 80
    critical: 95
  temperature:
    warning: 180
    critical: 200
emergency_actions:
  - condition: "pressure > 95 OR temperature > 200"
    action: "emergency_shutdown"
    notification: "immediate"
```

## Security Considerations

- Sensor data is encrypted in transit and at rest using industry-standard encryption
- Access control ensures only authorized agents can access specific sensor data
- Anomaly detection includes security threat identification (unauthorized sensor access, data tampering)
- Audit logging tracks all sensor data access and processing activities for compliance
- Edge processing reduces data transmission and potential attack surface

## Configuration

The sensor-integration skill can be configured with the following parameters:

- `default_protocols`: Default sensor protocols (mqtt, http, coap, bluetooth_le)
- `data_retention_period`: Data retention period for historical analysis (default: 90 days)
- `alert_severity_levels`: Alert severity levels and corresponding actions
- `ml_models_enabled`: Enable machine learning models for predictive analytics (default: true)
- `edge_processing_enabled`: Enable edge processing for low-latency responses (default: true)

This skill is essential for any agent that needs to monitor environmental conditions, implement predictive maintenance, manage smart agriculture systems, or ensure industrial safety through comprehensive sensor integration and analysis.