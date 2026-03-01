#!/bin/bash
set -euo pipefail

BACKUP_DIR="/tmp/briefcast-backup"
RCLONE_REMOTE="${RCLONE_REMOTE:-yandex}"
RCLONE_BUCKET="${RCLONE_BUCKET:-briefcast-backup}"
DB_PATH="/opt/briefcast/data/briefcast.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup SQLite using .backup command (safe for WAL mode)
if [ -f "$DB_PATH" ]; then
    sqlite3 "$DB_PATH" ".backup '$BACKUP_DIR/briefcast_$DATE.db'"
    gzip "$BACKUP_DIR/briefcast_$DATE.db"
    
    # Upload to cloud
    rclone copy "$BACKUP_DIR/" "$RCLONE_REMOTE:$RCLONE_BUCKET/" --include "*.gz"
    
    # Keep only last 30 backups remotely
    rclone delete "$RCLONE_REMOTE:$RCLONE_BUCKET/" --min-age 30d
    
    echo "Backup completed: briefcast_$DATE.db.gz"
else
    echo "Database not found at $DB_PATH"
    exit 1
fi

# Cleanup
rm -rf "$BACKUP_DIR"
