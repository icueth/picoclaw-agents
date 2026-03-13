---
name: device-control
description: Comprehensive smart home and IoT device control system for AI agents with multi-protocol support and automation capabilities
---

# Device Control

This built-in skill provides comprehensive smart home and IoT device control capabilities for AI agents to manage, control, and automate various smart devices and IoT systems across multiple protocols and platforms.

## Capabilities

- **Multi-Protocol Support**: Support for MQTT, HTTP/HTTPS, WebSocket, Zigbee, Z-Wave, Bluetooth LE, and proprietary APIs
- **Device Integration**: Integrate with popular smart home platforms (Home Assistant, SmartThings, Hubitat, OpenHAB, Apple HomeKit, Google Home, Amazon Alexa)
- **Device Control**: Control lights, thermostats, locks, cameras, speakers, appliances, and sensors with precise commands
- **Scene Management**: Create and manage complex scenes that control multiple devices simultaneously
- **Automation Rules**: Create automation rules based on time, sensor data, device states, and external triggers
- **Voice Integration**: Integrate with voice assistants for hands-free control and natural language commands
- **Energy Management**: Monitor and optimize energy usage across connected devices and appliances
- **Security Monitoring**: Monitor security devices (cameras, door/window sensors, motion detectors) and trigger alerts
- **Remote Access**: Control devices remotely with secure authentication and encrypted communication
- **Device Discovery**: Automatically discover and configure new devices on the network

## Usage Examples

### Multi-Device Control
```yaml
tool: device-control
action: control_devices
devices:
  - type: "light"
    name: "Living Room Lights"
    platform: "home_assistant"
    command: "turn_on"
    parameters:
      brightness: 75
      color_temp: 3000
  - type: "thermostat"
    name: "Nest Thermostat"
    platform: "google_home"
    command: "set_temperature"
    parameters:
      temperature: 22
      mode: "heat"
  - type: "speaker"
    name: "Sonos Living Room"
    platform: "sonos"
    command: "play_playlist"
    parameters:
      playlist: "Relaxing Evening"
      volume: 30
```

### Scene Management
```yaml
tool: device-control
action: create_scene
scene_name: "Movie Night"
devices:
  - type: "light"
    name: "Living Room Main"
    command: "turn_off"
  - type: "light"
    name: "Living Room Accent"
    command: "turn_on"
    parameters:
      brightness: 20
      color: "warm_white"
  - type: "tv"
    name: "Samsung TV"
    command: "turn_on"
    parameters:
      input: "HDMI1"
  - type: "speaker"
    name: "Soundbar"
    command: "turn_on"
    parameters:
      volume: 40
  - type: "blinds"
    name: "Living Room Blinds"
    command: "close"
```

### Automation Rules
```yaml
tool: device-control
action: create_automation
automation_name: "Morning Routine"
triggers:
  - type: "time"
    time: "07:00"
    days: ["monday", "tuesday", "wednesday", "thursday", "friday"]
  - type: "sensor"
    device: "Bedroom Motion Sensor"
    condition: "detected"
conditions:
  - type: "time_range"
    start: "06:00"
    end: "09:00"
actions:
  - type: "device"
    device: "Bedroom Lights"
    command: "turn_on"
    parameters:
      brightness: 50
  - type: "device"
    device: "Coffee Maker"
    command: "brew"
  - type: "device"
    device: "Thermostat"
    command: "set_temperature"
    parameters:
      temperature: 21
```

### Security Monitoring
```yaml
tool: device-control
action: monitor_security
sensors:
  - type: "motion"
    name: "Front Door Motion"
    zone: "perimeter"
  - type: "door"
    name: "Front Door Contact"
    zone: "entry"
  - type: "camera"
    name: "Front Door Camera"
    zone: "perimeter"
alerts:
  - condition: "motion_detected AND door_closed"
    action: "send_notification"
    recipients: ["phone", "email"]
    priority: "high"
  - condition: "door_opened AND time NOT BETWEEN 06:00 AND 22:00"
    action: "trigger_alarm"
    duration: "30s"
```

## Security Considerations

- Device control commands are authenticated and authorized before execution
- Communication with devices uses encrypted protocols (TLS, WPA3) where supported
- Access control ensures only authorized agents can control specific devices
- Audit logging tracks all device control activities for security monitoring
- Sensitive device credentials are securely stored and never exposed in logs

## Configuration

The device-control skill can be configured with the following parameters:

- `default_platform`: Default smart home platform (home_assistant, smartthings, openhab)
- `security_level`: Security level for device control (basic, standard, high)
- `notification_channels`: Enabled notification channels (push, email, sms, webhook)
- `automation_timeout`: Maximum duration for automation execution (default: 5m)
- `device_discovery_enabled`: Enable automatic device discovery (default: true)

This skill is essential for any agent that needs to control smart home devices, create automation routines, monitor security systems, or integrate IoT devices into broader automation workflows.