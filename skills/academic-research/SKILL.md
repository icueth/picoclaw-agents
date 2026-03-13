---
name: academic-research
description: Comprehensive academic research assistant with paper search, analysis, and citation management capabilities
---

# Academic Research

This built-in skill provides comprehensive academic research capabilities for AI agents to search, analyze, and manage academic papers, citations, and research data.

## Capabilities

- **Paper Search**: Search academic databases (arXiv, PubMed, Semantic Scholar, CrossRef, Google Scholar)
- **Paper Analysis**: Extract key information, summarize content, and analyze research quality
- **Citation Management**: Manage citations in various formats (APA, MLA, Chicago, IEEE, BibTeX)
- **Literature Review**: Conduct systematic literature reviews and meta-analyses
- **Research Trend Analysis**: Identify emerging trends and influential papers in specific fields
- **Author Tracking**: Track specific authors and their publication history
- **Journal Analysis**: Analyze journal impact factors, acceptance rates, and publication patterns
- **Reference Extraction**: Extract references from PDFs and other document formats
- **Collaboration Discovery**: Identify potential research collaborators based on publication patterns
- **Data Repository Integration**: Integrate with research data repositories (Figshare, Zenodo, Dryad)

## Usage Examples

### Search Papers
```yaml
tool: academic-research
action: search_papers
query: "large language models for code generation"
databases: ["arxiv", "semantic_scholar"]
date_range:
  start: "2023-01-01"
  end: "2026-03-13"
max_results: 20
sort_by: "relevance"
```

### Analyze Paper
```yaml
tool: academic-research
action: analyze_paper
paper_id: "arxiv:2305.12345"
analysis_type: "comprehensive"
include:
  - "summary"
  - "methodology"
  - "results"
  - "limitations"
  - "citations"
```

### Generate Citations
```yaml
tool: academic-research
action: generate_citations
papers:
  - "arxiv:2305.12345"
  - "doi:10.1038/s41586-023-12345-6"
format: "apa"
include_doi: true
include_urls: true
```

## Security Considerations

- Research data is handled according to academic integrity standards
- Copyright compliance ensures proper use of published materials
- Privacy protection for sensitive research data and unpublished work
- Secure authentication for accessing subscription-based academic databases
- Audit logging tracks all research activities for reproducibility

## Configuration

The academic-research skill can be configured with the following parameters:

- `default_databases`: Default academic databases to search (default: arxiv, semantic_scholar)
- `citation_format`: Default citation format (default: apa)
- `max_papers_per_search`: Maximum papers to return per search (default: 50)
- `api_keys`: API keys for premium academic database access
- `cache_duration`: Duration to cache search results (default: 7 days)
- `export_formats`: Supported export formats (bibtex, ris, csv, json)

This skill is essential for any agent that needs to conduct academic research, write scholarly papers, or stay current with scientific literature. It provides comprehensive research support while maintaining academic integrity and copyright compliance.