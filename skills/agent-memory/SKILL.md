---
name: agent-memory
description: Persistent memory system for AI agents with semantic search and context management
---

# Agent Memory

This built-in skill provides a persistent memory system for AI agents to store, retrieve, and manage contextual information across sessions and interactions.

## Capabilities

- **Memory Storage**: Store structured and unstructured data with metadata
- **Semantic Search**: Retrieve memories using natural language queries with vector similarity
- **Context Management**: Maintain conversation context and relevant memories
- **Memory Compression**: Compress and summarize long-term memories for efficiency
- **Temporal Awareness**: Track when memories were created and accessed
- **Memory Linking**: Create relationships between related memories
- **Privacy Controls**: Manage memory access permissions and data retention
- **Export/Import**: Export memories to various formats and import from external sources
- **Memory Cleanup**: Automatically clean up outdated or irrelevant memories
- **Multi-modal Support**: Store and retrieve text, images, and other data types

## Usage Examples

### Store Memory
```yaml
tool: agent-memory
action: store
memory:
  content: "User prefers dark mode and uses VS Code as their primary editor"
  tags: ["user_preferences", "development"]
  context: "conversation_12345"
  importance: 0.8
```

### Retrieve Memory
```yaml
tool: agent-memory
action: retrieve
query: "What are the user's development preferences?"
context: "current_conversation"
limit: 5
```

### Update Memory
```yaml
tool: agent-memory
action: update
memory_id: "mem_67890"
updates:
  content: "User now prefers light mode during daytime hours"
  tags: ["user_preferences", "development", "ui_theme"]
```

## Security Considerations

- All memory data is encrypted at rest using industry-standard encryption
- Access control ensures only authorized agents can access specific memories
- Data retention policies automatically expire sensitive information
- Privacy-by-design principles prevent unauthorized data collection
- Audit logging tracks all memory operations for compliance

## Configuration

The agent-memory skill can be configured with the following parameters:

- `storage_backend`: Storage backend (sqlite, postgres, memory)
- `embedding_model`: Embedding model for semantic search (default: sentence-transformers)
- `max_memory_size`: Maximum size of individual memories (default: 10KB)
- `retention_policy`: Data retention policy (default: 30 days for non-essential data)
- `privacy_level`: Privacy level controls (strict, moderate, relaxed)
- `compression_enabled`: Enable memory compression (default: true)

This skill is essential for any agent that needs to maintain context across interactions, learn from past experiences, or provide personalized responses based on historical data.