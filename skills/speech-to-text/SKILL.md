---
name: speech-to-text
description: Advanced speech-to-text transcription system for AI agents with multi-language support and real-time processing capabilities
---

# Speech-to-Text

This built-in skill provides advanced speech-to-text transcription capabilities for AI agents to convert spoken audio into accurate text with support for multiple languages, speakers, and audio formats.

## Capabilities

- **Multi-Language Support**: Transcribe speech in over 100 languages including Thai, Chinese, Japanese, Arabic, and European languages
- **Speaker Diarization**: Identify and separate different speakers in multi-person conversations
- **Real-Time Transcription**: Process live audio streams with low latency for real-time applications
- **Audio Format Support**: Handle various audio formats (WAV, MP3, FLAC, OGG, M4A, WebM)
- **Noise Reduction**: Automatically reduce background noise and enhance speech clarity
- **Punctuation and Formatting**: Add appropriate punctuation, capitalization, and formatting to transcribed text
- **Custom Vocabulary**: Use custom vocabulary and domain-specific terms for improved accuracy
- **Confidence Scoring**: Provide confidence scores for transcribed words and phrases
- **Batch Processing**: Process multiple audio files simultaneously with consistent results
- **Streaming API**: Support streaming audio input for long-form content and live transcription

## Usage Examples

### Basic Audio Transcription
```yaml
tool: speech-to-text
action: transcribe_audio
audio_path: "/recordings/meeting.mp3"
language: "en-US"
speaker_diarization: true
punctuation: true
output_format: "text"
confidence_threshold: 0.8
```

### Real-Time Streaming Transcription
```yaml
tool: speech-to-text
action: stream_transcription
stream_url: "rtmp://live.example.com/stream"
language: "th-TH"
real_time: true
speaker_diarization: true
output_format: "json"
include_timestamps: true
```

### Multi-Language Transcription
```yaml
tool: speech-to-text
action: transcribe_multilingual
audio_path: "/recordings/international_call.wav"
languages: ["en-US", "th-TH", "zh-CN", "ja-JP"]
auto_detect_language: true
speaker_diarization: true
output_format: "srt"
include_confidence: true
```

### Custom Vocabulary Enhancement
```yaml
tool: speech-to-text
action: transcribe_with_vocabulary
audio_path: "/recordings/tech_meeting.wav"
language: "en-US"
custom_vocabulary:
  - "OpenClaw"
  - "PicoClaw"
  - "A2A orchestration"
  - "built-in skills"
  - "agent-browser"
domain: "technology"
output_format: "text"
```

## Security Considerations

- Audio processing runs in isolated environments to prevent data leakage
- Sensitive audio content is encrypted at rest and never transmitted to external services
- Access control ensures only authorized agents can process specific audio files
- Audit logging tracks all transcription activities for compliance and security monitoring
- Content filtering prevents processing of potentially inappropriate or malicious audio content

## Configuration

The speech-to-text skill can be configured with the following parameters:

- `default_language`: Default language for transcription (default: en-US)
- `engine_backend`: Transcription engine backend (whisper, assemblyai, google_speech, custom)
- `max_audio_duration`: Maximum audio duration for processing (default: 1h)
- `temp_directory`: Temporary directory for audio processing (default: system temp)
- `privacy_mode`: Privacy mode for handling sensitive audio (strict, moderate, relaxed)

This skill is essential for any agent that needs to transcribe audio recordings, process live speech, handle multilingual content, or convert spoken communication into searchable and analyzable text.