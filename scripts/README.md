# Sentinel Deployment Scripts

This directory contains scripts for automated installation and upgrades of Sentinel monitoring system.

## Scripts Overview

### install.sh

Automated installation script for Linux systems with systemd support.

**Features:**

- Platform detection (Linux amd64/arm64)
- Automatic binary download from GitHub releases
- System user creation with security hardening
- Directory structure setup with proper permissions
- Systemd service configuration with security features
- Configuration file creation with sensible defaults
- Service startup and health verification

**Usage:**

```bash
# Basic installation
curl -L https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/install.sh | sudo bash

# Custom installation directories
sudo ./install.sh --install-dir /opt/sentinel/bin --config-dir /etc/sentinel

# Install specific version
sudo ./install.sh --version v1.0.0

# Custom service configuration
sudo ./install.sh --service-name sentinel-monitor --user sentinel-svc
```

**Options:**

- `-d, --install-dir DIR`: Binary installation directory (default: `/usr/local/bin`)
- `-c, --config-dir DIR`: Configuration directory (default: `/etc/sentinel`)
- `-D, --data-dir DIR`: Data directory (default: `/var/lib/sentinel`)
- `-s, --service-name NAME`: Systemd service name (default: `sentinel`)
- `-u, --user USER`: Service user (default: `sentinel`)
- `-v, --version VERSION`: Install specific version (default: latest)
- `-h, --help`: Show help message

**Post-Installation:**

- Service: `systemctl status sentinel`
- Configuration: `/etc/sentinel/config.yaml`
- Logs: `journalctl -u sentinel -f`
- Web interface: <http://localhost:8080>

### upgrade.sh

Automated upgrade script for existing Sentinel installations.

**Features:**

- Automatic latest version detection
- Service backup before upgrade
- Graceful service shutdown and restart
- Configuration preservation
- Health checks after upgrade
- Automatic rollback on failure
- Minimal downtime (10-30 seconds)

**Usage:**

```bash
# Upgrade to latest version
curl -L https://raw.githubusercontent.com/sxwebdev/sentinel/master/scripts/upgrade.sh | sudo bash

# Upgrade specific service
sudo ./upgrade.sh sentinel-monitor

# Upgrade to specific version
sudo ./upgrade.sh --version v1.0.0

# Dry run (check what would be updated)
sudo ./upgrade.sh --dry-run

# Force upgrade (skip confirmations)
sudo ./upgrade.sh --force
```

**Options:**

- `SERVICE_NAME`: Target service name (positional argument, default: `sentinel`)
- `--version VERSION`: Upgrade to specific version
- `--dry-run`: Show what would be upgraded without making changes
- `--force`: Skip confirmation prompts
- `--help`: Show help message

## Requirements

### System Requirements

- Linux operating system (Ubuntu, CentOS, RHEL, Debian, etc.)
- systemd init system
- Root or sudo privileges
- Internet connection for downloads

### Dependencies

Both scripts automatically check for and require these tools:

- `curl`: For downloading files
- `jq`: For JSON processing
- `tar`: For archive extraction
- `systemctl`: For service management

### Platform Support

- **Architecture**: amd64 (x86_64), arm64 (aarch64)
- **Operating System**: Linux distributions with systemd
- **Not Supported**: macOS (no systemd), Windows

## Best Practices

### Production Deployment

1. **Enable Authentication**: Set `auth.enabled: true` and use strong passwords
1. **Use HTTPS**: Configure reverse proxy (nginx, Apache) with SSL/TLS
1. **Regular Backups**: Backup configuration and database regularly
1. **Monitor Resources**: Set up monitoring for the Sentinel service itself
1. **Log Rotation**: Configure logrotate for journal logs

### Maintenance

1. **Regular Updates**: Use upgrade script monthly or when security updates are available
1. **Health Monitoring**: Include Sentinel service in your monitoring stack
1. **Configuration Review**: Periodically review and update configurations
1. **Performance Tuning**: Adjust intervals and timeouts based on your needs

### Security

1. **User Permissions**: Never run Sentinel as root in production
1. **Network Security**: Use firewall rules to restrict access
1. **Regular Audits**: Review service logs and access patterns
1. **Update Management**: Apply security updates promptly

## License

These scripts are part of the Sentinel project and are licensed under the MIT License.
