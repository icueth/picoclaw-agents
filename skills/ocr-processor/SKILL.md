---
name: ocr-processor
description: Advanced optical character recognition system for AI agents with multi-language support and document intelligence
---

# OCR Processor

This built-in skill provides advanced optical character recognition (OCR) capabilities for AI agents to extract text from images, scanned documents, and PDF files with high accuracy and multi-language support.

## Capabilities

- **Multi-Language OCR**: Recognize text in over 100 languages including Thai, Chinese, Japanese, Arabic, and European languages
- **Document Layout Analysis**: Preserve document structure, tables, columns, and formatting during text extraction
- **Handwriting Recognition**: Recognize handwritten text with specialized models for different writing styles
- **Image Preprocessing**: Automatically enhance image quality, remove noise, and correct skew before OCR processing
- **Batch Processing**: Process multiple images or documents simultaneously with consistent results
- **Confidence Scoring**: Provide confidence scores for extracted text to identify low-quality results
- **Region Detection**: Detect and extract text from specific regions or areas of interest in images
- **Barcode/QR Code Reading**: Extract data from barcodes, QR codes, and other machine-readable symbols
- **Document Classification**: Automatically classify document types and apply appropriate OCR settings
- **Output Formats**: Generate output in various formats (plain text, JSON, XML, searchable PDF, hOCR)

## Usage Examples

### Basic OCR on Image
```yaml
tool: ocr-processor
action: extract_text
image_path: "/images/receipt.jpg"
language: "eng+tha"
output_format: "text"
confidence_threshold: 0.8
```

### Document Layout Preservation
```yaml
tool: ocr-processor
action: extract_with_layout
document_path: "/scans/multi_column_article.pdf"
language: "eng"
preserve_tables: true
preserve_columns: true
output_format: "markdown"
```

### Handwriting Recognition
```yaml
tool: ocr-processor
action: recognize_handwriting
image_path: "/images/notes.jpg"
handwriting_model: "general"
language: "eng"
output_format: "json"
include_confidence: true
```

### Batch OCR Processing
```yaml
tool: ocr-processor
action: batch_process
input_directory: "/scans/invoices/"
output_directory: "/extracted_texts/"
language: "eng+fra+deu"
file_pattern: "*.jpg"
output_format: "searchable_pdf"
```

## Security Considerations

- All OCR processing runs in isolated environments to prevent data leakage
- Sensitive document data is encrypted at rest and never transmitted to external services
- Access control ensures only authorized agents can process specific documents
- Audit logging tracks all OCR operations for compliance and security monitoring
- Content filtering prevents processing of potentially malicious or inappropriate content

## Configuration

The ocr-processor skill can be configured with the following parameters:

- `default_language`: Default language for OCR processing (default: eng)
- `engine_backend`: OCR engine backend (tesseract, easyocr, paddleocr, custom)
- `max_image_size`: Maximum image size for processing (default: 10MB)
- `temp_directory`: Temporary directory for processing (default: system temp)
- `quality_enhancement`: Enable automatic image quality enhancement (default: true)
- `privacy_mode`: Privacy mode for handling sensitive documents (strict, moderate, relaxed)

This skill is essential for any agent that needs to extract text from images, process scanned documents, recognize handwriting, or convert physical documents into digital, searchable formats.