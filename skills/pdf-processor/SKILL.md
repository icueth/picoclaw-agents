---
name: pdf-processor
description: Comprehensive PDF processing and manipulation system for AI agents with advanced document capabilities
---

# PDF Processor

This built-in skill provides comprehensive PDF processing and manipulation capabilities for AI agents to handle document workflows, extract information, and generate professional PDF documents.

## Capabilities

- **PDF Creation**: Generate PDF documents from HTML, Markdown, text, or structured data
- **PDF Extraction**: Extract text, images, tables, and metadata from PDF documents
- **PDF Manipulation**: Merge, split, rotate, crop, and reorganize PDF pages
- **Watermarking**: Add text or image watermarks to PDF documents
- **OCR Processing**: Perform optical character recognition on scanned PDFs and images
- **Form Handling**: Fill, extract, and validate PDF forms and interactive elements
- **Security Features**: Password protection, encryption, and digital signatures
- **Compression**: Optimize PDF file size while maintaining quality
- **Metadata Management**: Read, write, and manage PDF metadata and properties
- **Batch Processing**: Process multiple PDF files simultaneously with consistent formatting

## Usage Examples

### Create PDF from HTML
```yaml
tool: pdf-processor
action: create_pdf
source:
  type: "html"
  content: "<h1>Report Title</h1><p>Report content...</p>"
output_path: "/reports/monthly_report.pdf"
options:
  page_size: "A4"
  margins: "1in"
  header: "Monthly Report - {{date}}"
```

### Extract Text from PDF
```yaml
tool: pdf-processor
action: extract_text
pdf_path: "/documents/contract.pdf"
output_format: "structured"
include_metadata: true
extract_tables: true
```

### Add Watermark
```yaml
tool: pdf-processor
action: add_watermark
pdf_path: "/reports/confidential.pdf"
watermark:
  text: "CONFIDENTIAL"
  opacity: 0.3
  rotation: 45
  font_size: 50
  color: "#FF0000"
output_path: "/reports/confidential_watermarked.pdf"
```

### OCR Processing
```yaml
tool: pdf-processor
action: ocr_process
pdf_path: "/scans/invoice_scan.pdf"
language: "eng+tha"
output_format: "searchable_pdf"
preserve_layout: true
```

## Security Considerations

- All PDF processing runs in isolated sandboxed environments
- Sensitive document data is encrypted at rest and in transit
- Access control ensures only authorized agents can process specific documents
- Audit logging tracks all PDF operations for compliance and security
- Malicious content detection prevents exploitation of PDF vulnerabilities

## Configuration

The pdf-processor skill can be configured with the following parameters:

- `default_engine`: Default PDF engine (pdfkit, weasyprint, puppeteer)
- `ocr_enabled`: Enable OCR processing (default: true)
- `max_file_size`: Maximum file size for processing (default: 50MB)
- `temp_directory`: Temporary directory for processing (default: system temp)
- `security_level`: Security level for PDF processing (strict, moderate, relaxed)

This skill is essential for any agent that needs to handle document workflows, generate reports, extract information from PDFs, or perform professional document processing tasks.