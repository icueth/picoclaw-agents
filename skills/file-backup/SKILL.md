---
name: file-backup
description: Comprehensive file backup and recovery system for AI agents with encryption, versioning, and cloud storage support
---

# File Backup

This built-in skill provides comprehensive file backup and recovery capabilities for AI agents to protect data, ensure availability, and maintain version history across multiple storage backends.

## Capabilities

- **Incremental Backups**: Perform efficient incremental backups to minimize storage and bandwidth usage
- **Versioning**: Maintain version history with point-in-time recovery capabilities
- **Encryption**: Encrypt backup data at rest and in transit using strong encryption algorithms
- **Cloud Storage Integration**: Support multiple cloud storage providers (AWS S3, Google Cloud, Azure, Dropbox, etc.)
- **Local Storage**: Support local and network-attached storage for backup destinations
- **Compression**: Apply compression to reduce backup size and improve transfer efficiency
- **Scheduling**: Automate backup schedules with flexible timing options
- **Verification**: Verify backup integrity and perform regular restore tests
- **Retention Policies**: Implement configurable retention policies for automatic cleanup
- **Monitoring and Alerts**: Monitor backup status and send alerts for failures or issues

## Usage Examples

### Create Backup
```yaml
tool: file-backup
action: create_backup
source:
  paths:
    - "/home/user/documents"
    - "/home/user/projects"
  exclude_patterns:
    - "*.tmp"
    - "node_modules/"
destination:
  type: "s3"
  bucket: "user-backups"
  path: "daily/{{timestamp}}"
encryption:
  enabled: true
  key_id: "backup-key-2026"
compression: "gzip"
incremental: true
```

### Restore Files
```yaml
tool: file-backup
action: restore_backup
backup_id: "backup-2026-03-12-1430"
destination: "/home/user/restore"
version: "latest"
verify_integrity: true
```

### Schedule Backup
```yaml
tool: file-backup
action: schedule_backup
schedule:
  type: "cron"
  expression: "0 2 * * *"  # Daily at 2 AM
backup_config:
  source:
    paths: ["/home/user"]
  destination:
    type: "google_cloud"
    bucket: "user-backups"
  retention:
    daily: 7
    weekly: 4
    monthly: 12
```

## Security Considerations

- All backup data is encrypted using AES-256 encryption at rest and TLS in transit
- Encryption keys are managed securely using hardware security modules (HSM) or key management services
- Access control ensures only authorized agents can create, modify, or restore backups
- Audit logging tracks all backup operations for compliance and security monitoring
- Secure credential management handles cloud storage authentication tokens safely

## Configuration

The file-backup skill can be configured with the following parameters:

- `default_encryption`: Enable encryption by default (default: true)
- `default_compression`: Default compression algorithm (gzip, bzip2, lz4, zstd)
- `max_backup_size`: Maximum size for individual backups (default: 10GB)
- `retention_policy`: Default retention policy (daily: 7, weekly: 4, monthly: 12)
- `verification_frequency`: Frequency of backup verification tests (default: weekly)
- `storage_providers`: Enabled cloud storage provider integrations

This skill is essential for any agent that needs to protect important data, ensure business continuity, or maintain version history of critical files. It provides robust backup and recovery capabilities while maintaining data security and compliance.