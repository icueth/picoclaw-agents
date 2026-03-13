---
name: personal-knowledge-management
description: Advanced personal knowledge management and note-taking system for AI agents with intelligent organization, retrieval, and synthesis capabilities
---

# Personal Knowledge Management

This built-in skill provides advanced personal knowledge management and note-taking capabilities for AI agents to capture, organize, retrieve, and synthesize personal knowledge through intelligent systems and structured methodologies.

## Capabilities

- **Note Capture**: Capture notes from various sources (text, voice, web, documents, conversations) with automatic formatting and metadata
- **Intelligent Organization**: Organize notes using multiple systems (Zettelkasten, PARA, MOC, folders, tags, links) with automatic categorization
- **Semantic Search**: Retrieve notes using natural language queries with semantic understanding and context awareness
- **Knowledge Synthesis**: Synthesize information from multiple notes to create new insights, summaries, and connections
- **Linking and Connections**: Create bidirectional links between related notes with automatic relationship detection
- **Template System**: Use templates for consistent note structure and content types (meeting notes, book notes, research notes)
- **Integration Hub**: Integrate with popular note-taking tools (Obsidian, Notion, Roam Research, Logseq, Evernote)
- **Version Control**: Track note versions and changes with full history and rollback capabilities
- **Analytics and Insights**: Provide knowledge management analytics with usage patterns, connection insights, and gap analysis
- **Privacy and Security**: Ensure privacy and security of personal knowledge with encryption and access controls

## Usage Examples

### Intelligent Note Capture
```yaml
tool: personal-knowledge-management
action: capture_note
source: "meeting_transcript"
content: |
  Meeting with Product Team - March 13, 2026
  - Discussed Q2 roadmap priorities
  - Decided to focus on user onboarding improvements
  - Engineering bandwidth is constrained, need to prioritize
  - Marketing will support with beta tester recruitment
metadata:
  date: "2026-03-13"
  participants: ["john", "sarah", "mike"]
  project: "product_alpha"
  tags: ["meeting", "roadmap", "priorities"]
template: "meeting_notes"
auto_link: true
```

### Semantic Search and Retrieval
```yaml
tool: personal-knowledge-management
action: search_notes
query: "What were our Q2 priorities for product alpha?"
search_types:
  - "semantic"
  - "keyword"
  - "tag_based"
  - "date_range"
date_range:
  start: "2026-01-01"
  end: "2026-03-31"
tags: ["product_alpha", "roadmap"]
context_awareness: true
relevance_threshold: 0.7
```

### Knowledge Synthesis
```yaml
tool: personal-knowledge-management
action: synthesize_knowledge
topic: "Product Alpha User Onboarding Strategy"
source_notes:
  - "meeting_notes_2026-03-13"
  - "user_research_findings_q1"
  - "competitor_analysis_onboarding"
  - "technical_constraints_document"
synthesis_type: "comprehensive_summary"
output_format: "structured_document"
include_connections: true
highlight_gaps: true
```

### Zettelkasten Organization
```yaml
tool: personal-knowledge-management
action: organize_zettelkasten
notes:
  - id: "202603131430"
    title: "Product Alpha Q2 Priorities"
    content: "Focus on user onboarding improvements..."
    tags: ["product_alpha", "roadmap", "onboarding"]
    links: ["202602151200", "202601200900"]
  - id: "202603131431"
    title: "User Onboarding Pain Points"
    content: "Users struggle with account setup and initial configuration..."
    tags: ["user_research", "onboarding", "ux"]
    links: ["202603131430", "202602281500"]
structure_type: "zettelkasten"
auto_categorize: true
detect_relationships: true
```

## Security Considerations

- Personal knowledge data is encrypted at rest and in transit using industry-standard encryption
- Access control ensures only authorized agents can access specific knowledge information
- Privacy controls allow users to manage what knowledge data is shared and with whom
- Audit logging tracks all knowledge management activities for accountability and security
- Sensitive personal information is automatically detected and protected with appropriate access controls

## Configuration

The personal-knowledge-management skill can be configured with the following parameters:

- `default_organization_system`: Default organization system (zettelkasten, para, moc, folders)
- `semantic_search_enabled`: Enable semantic search by default (default: true)
- `auto_linking_level`: Level of automatic linking (none, basic, comprehensive)
- `integration_platforms`: Enabled integration platforms (obsidian, notion, roam, logseq)
- `privacy_level`: Privacy level for knowledge data sharing (private, selective, public)

This skill is essential for any agent that needs to help users capture and organize personal knowledge, retrieve information intelligently, synthesize insights from multiple sources, or provide advanced personal knowledge management and note-taking capabilities.