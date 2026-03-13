---
name: text-to-speech
description: Advanced text-to-speech synthesis system for AI agents with multi-language support and natural voice generation capabilities
---

# Text-to-Speech

This built-in skill provides advanced text-to-speech synthesis capabilities for AI agents to convert text into natural-sounding speech with support for multiple languages, voices, and audio formats.

## Capabilities

- **Multi-Language Support**: Generate speech in over 100 languages including Thai, Chinese, Japanese, Arabic, and European languages
- **Voice Variety**: Choose from multiple natural-sounding voices with different characteristics (gender, age, accent, emotion)
- **Emotion and Prosody**: Control speech emotion, intonation, pitch, speed, and emphasis for natural delivery
- **Audio Format Support**: Generate audio in various formats (WAV, MP3, FLAC, OGG, M4A)
- **SSML Support**: Use Speech Synthesis Markup Language (SSML) for advanced speech control
- **Batch Processing**: Generate multiple audio files simultaneously with consistent quality
- **Real-Time Streaming**: Stream audio output in real-time for interactive applications
- **Custom Voice Cloning**: Create custom voice models from reference audio samples (with proper consent)
- **Audio Effects**: Apply audio effects like background noise, reverb, or compression for specific contexts
- **Accessibility Features**: Support accessibility features like screen reader compatibility and audio descriptions

## Usage Examples

### Basic Text-to-Speech
```yaml
tool: text-to-speech
action: synthesize_speech
text: "Hello, welcome to our AI-powered assistant!"
language: "en-US"
voice: "en-US-Neural2-F"
output_format: "mp3"
output_path: "/audio/welcome.mp3"
```

### Multi-Language Synthesis
```yaml
tool: text-to-speech
action: synthesize_multilingual
text: "สวัสดี! Welcome to our AI platform. ยินดีต้อนรับสู่แพลตฟอร์ม AI ของเรา"
languages: ["th-TH", "en-US"]
voices:
  th-TH: "th-TH-Neural2-S"
  en-US: "en-US-Neural2-F"
output_format: "wav"
output_path: "/audio/multilingual_greeting.wav"
```

### SSML Enhanced Speech
```yaml
tool: text-to-speech
action: synthesize_with_ssml
ssml: |
  <speak>
    <prosody rate="slow" pitch="+10%">Welcome</prosody>
    to our <emphasis level="strong">AI-powered</emphasis> platform!
    <break time="500ms"/>
    How can I help you today?
  </speak>
language: "en-US"
voice: "en-US-Neural2-F"
output_format: "mp3"
```

### Real-Time Streaming
```yaml
tool: text-to-speech
action: stream_speech
text: "Processing your request. Please wait while I analyze the data..."
language: "en-US"
voice: "en-US-Neural2-F"
stream_type: "websocket"
stream_url: "ws://localhost:8080/audio"
real_time: true
```

## Security Considerations

- Voice synthesis runs in isolated environments to prevent unauthorized access
- Custom voice models require explicit consent and are securely stored
- Access control ensures only authorized agents can generate speech content
- Audit logging tracks all text-to-speech activities for compliance and security monitoring
- Content filtering prevents generation of inappropriate or harmful speech content

## Configuration

The text-to-speech skill can be configured with the following parameters:

- `default_language`: Default language for speech synthesis (default: en-US)
- `default_voice`: Default voice for each language
- `engine_backend`: TTS engine backend (google_tts, amazon_polly, azure_tts, custom)
- `max_text_length`: Maximum text length for synthesis (default: 5000 characters)
- `audio_quality`: Audio quality settings (low, medium, high, premium)

This skill is essential for any agent that needs to generate spoken content, create audio interfaces, support accessibility features, or provide natural voice interactions with users.