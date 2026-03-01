#!/bin/bash
set -euo pipefail

echo "=== Briefcast VPS Bootstrap ==="

# Update system
apt-get update && apt-get upgrade -y

# Install Docker
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
fi

# Install Docker Compose plugin
if ! docker compose version &> /dev/null; then
    apt-get install -y docker-compose-plugin
fi

# Install rclone for backups
if ! command -v rclone &> /dev/null; then
    curl https://rclone.org/install.sh | bash
fi

# Create app user
if ! id -u briefcast &> /dev/null 2>&1; then
    useradd -m -s /bin/bash briefcast
    usermod -aG docker briefcast
fi

# Create app directory
mkdir -p /opt/briefcast
chown briefcast:briefcast /opt/briefcast

# Setup backup cron (daily at 3am)
CRON_CMD="0 3 * * * /opt/briefcast/scripts/backup.sh >> /var/log/briefcast-backup.log 2>&1"
(crontab -u briefcast -l 2>/dev/null | grep -v backup.sh; echo "$CRON_CMD") | crontab -u briefcast -

echo "=== Bootstrap complete ==="
echo "Next steps:"
echo "1. Copy project files to /opt/briefcast/"
echo "2. Create .env file from .env.example"
echo "3. Configure rclone: rclone config"
echo "4. Run: cd /opt/briefcast && docker compose up -d"
