---
name: real-time-transcription
description: Real-time speech transcription and live captioning system for AI agents with low-latency processing and multi-speaker support
---

# Real-Time Transcription

This built-in skill provides real-time speech transcription and live captioning capabilities for AI agents to process live audio streams with low latency, supporting multiple speakers and dynamic environments.

## Capabilities

- **Live Audio Processing**: Process live audio streams from microphones, webcams, or network sources
- **Low-Latency Transcription**: Achieve sub-second latency for real-time applications and live captioning
- **Multi-Speaker Support**: Identify and transcribe multiple speakers simultaneously with speaker diarization
- **Dynamic Language Detection**: Automatically detect and switch between languages during transcription
- **Punctuation and Formatting**: Add real-time punctuation, capitalization, and formatting to live transcripts
- **Streaming Output**: Stream transcription results in real-time via WebSockets, HTTP streaming, or file output
- **Noise Adaptation**: Adapt to changing noise conditions and acoustic environments automatically
- **Vocabulary Adaptation**: Dynamically update vocabulary based on context and domain-specific terms
- **Error Correction**: Apply real-time error correction and confidence-based refinement
- **Integration APIs**: Provide APIs for integration with video conferencing, live streaming, and accessibility tools

## Usage Examples

### Live Meeting Transcription
```yaml
tool: real-time-transcription
action: start_live_transcription
audio_source:
  type: "microphone"
  device_id: "default"
languages: ["en-US", "th-TH"]
speaker_diarization: true
output_format: "json"
streaming_output:
  type: "websocket"
  url: "ws://localhost:8080/transcript"
punctuation: true
confidence_threshold: 0.7
```

### Video Conference Integration
```yaml
tool: real-time-transcription
action: integrate_with_video_conference
platform: "zoom"
meeting_id: "123456789"
languages: ["en-US"]
speaker_identification: true
caption_display: true
transcript_recording: true
output_formats:
  - "srt"
  - "vtt"
  - "json"
```

### Multi-Language Live Event
```yaml
tool: real-time-transcription
action: transcribe_live_event
audio_source:
  type: "network_stream"
  url: "rtmp://live.example.com/event"
languages: ["en-US", "th-TH", "zh-CN", "ja-JP"]
auto_language_detection: true
speaker_diarization: true
output_format: "srt"
streaming_output:
  type: "http_stream"
  url: "https://api.example.com/live-captions"
```

### Accessibility Integration
```yaml
tool: real-time-transcription
action: provide_accessibility_captions
audio_source:
  type: "system_audio"
  application: "video_player"
languages: ["en-US"]
caption_style:
  font_size: "large"
  background_color: "black"
  text_color: "white"
  position: "bottom"
real_time: true
low_latency: true
```

## Security Considerations

- **Real-Time Privacy**: Live audio is processed locally without transmission to external services
- **Access Control**: Only authorized agents can access live audio streams and transcription results
- **Data Minimization**: Only necessary audio data is processed and stored temporarily
- **Encryption**: Streaming outputs are encrypted using secure protocols (WSS, HTTPS)
- **Audit Logging**: All real-time transcription activities are logged for compliance and security monitoring
- **Content Filtering**: Real-time content filtering prevents inappropriate or harmful content

## Configuration

The real-time-transcription skill can be configured with the following parameters:

- `default_languages`: Default languages for real-time transcription
- `latency_target`: Target latency for real-time processing (ultra_low, low, standard)
- `max_speakers`: Maximum number of simultaneous speakers to track (default: 8)
- `audio_buffer_size`: Audio buffer size for latency vs. accuracy trade-off (default: 200ms)
- `privacy_mode`: Privacy mode for handling sensitive live audio (strict, moderate, relaxed)

This skill is essential for any agent that needs to provide live captioning, transcribe meetings in real-time, support accessibility features, or integrate speech recognition into live applications and events.