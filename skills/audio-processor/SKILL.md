---
name: audio-processor
description: Comprehensive audio processing and manipulation system for AI agents with advanced editing and enhancement capabilities
---

# Audio Processor

This built-in skill provides comprehensive audio processing and manipulation capabilities for AI agents to edit, enhance, and transform audio content with professional-grade quality and precision.

## Capabilities

- **Audio Format Conversion**: Convert between various audio formats (WAV, MP3, FLAC, OGG, M4A, WebM, AAC)
- **Noise Reduction**: Remove background noise, hum, hiss, and other unwanted audio artifacts
- **Audio Enhancement**: Enhance audio quality with equalization, compression, and normalization
- **Speech Isolation**: Isolate speech from background music or noise using AI-powered separation
- **Audio Editing**: Cut, trim, merge, split, and rearrange audio segments with precise control
- **Volume Control**: Adjust volume levels, normalize audio, and apply dynamic range compression
- **Speed and Pitch Manipulation**: Change playback speed and pitch independently without quality loss
- **Audio Effects**: Apply professional audio effects (reverb, echo, chorus, flanger, distortion)
- **Metadata Management**: Read, write, and manage audio metadata (ID3 tags, Vorbis comments)
- **Batch Processing**: Process multiple audio files simultaneously with consistent settings

## Usage Examples

### Noise Reduction and Enhancement
```yaml
tool: audio-processor
action: enhance_audio
audio_path: "/recordings/interview_raw.wav"
enhancements:
  - type: "noise_reduction"
    strength: 0.7
  - type: "equalization"
    preset: "voice_clarity"
  - type: "compression"
    ratio: 4.0
    threshold: -20
  - type: "normalization"
    target_level: -1.0
output_format: "wav"
output_path: "/recordings/interview_enhanced.wav"
```

### Audio Format Conversion
```yaml
tool: audio-processor
action: convert_format
audio_path: "/music/song.flac"
target_format: "mp3"
quality_settings:
  bitrate: 320
  sample_rate: 44100
  channels: 2
output_path: "/music/song.mp3"
preserve_metadata: true
```

### Speech Isolation
```yaml
tool: audio-processor
action: isolate_speech
audio_path: "/recordings/meeting_with_music.mp3"
isolation_strength: 0.9
background_music_suppression: 0.8
output_format: "wav"
speech_output: "/recordings/meeting_speech_only.wav"
background_output: "/recordings/meeting_background_only.wav"
```

### Audio Editing and Trimming
```yaml
tool: audio-processor
action: edit_audio
audio_path: "/podcasts/episode_full.mp3"
edits:
  - type: "trim"
    start_time: "00:05:30"
    end_time: "00:45:20"
  - type: "remove_section"
    start_time: "00:23:15"
    end_time: "00:23:45"
  - type: "fade_in"
    duration: "2s"
  - type: "fade_out"
    duration: "3s"
output_path: "/podcasts/episode_edited.mp3"
```

## Security Considerations

- Audio processing runs in isolated environments to prevent data leakage
- Sensitive audio content is encrypted at rest during processing
- Access control ensures only authorized agents can process specific audio files
- Audit logging tracks all audio processing activities for compliance and security monitoring
- Content filtering prevents processing of potentially inappropriate audio content

## Configuration

The audio-processor skill can be configured with the following parameters:

- `default_output_format`: Default output format (wav, mp3, flac, ogg)
- `max_file_size`: Maximum file size for processing (default: 100MB)
- `temp_directory`: Temporary directory for audio processing (default: system temp)
- `quality_preset`: Default quality preset (low, medium, high, premium)
- `privacy_mode`: Privacy mode for handling sensitive audio (strict, moderate, relaxed)

This skill is essential for any agent that needs to process audio recordings, enhance speech quality, convert audio formats, or perform professional-grade audio editing and manipulation.