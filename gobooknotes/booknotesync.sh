#!/bin/bash

. /home/efirlus/goproject/gobooknotes/logger.sh

# Paths
DB_FILE="/NAS/samba/Book/metadata.db"
MD_DIR="/home/efirlus/OneDrive/obsidian/Vault/6. Calibre"
LOG_FILE="/home/efirlus/goproject/Logs/app.log"

# Ignore times (10분 예정)
IGNORE_TIME=600

# Last event timestamps (initialized to 0)
last_db_event=0
last_md_event=0

log "info" "test5 initiated"
# Monitor events
inotifywait -m \
    --format '%e %w%f' \
    -e close_write "$DB_FILE" \
    -e attrib "$MD_DIR"/*.md | while read event file; do
    current_time=$(date +%s)
    
    if [[ "$file" == "$DB_FILE" && "$event" == *"CLOSE_WRITE"* ]]; then
        # Handle dbEvent
        if (( current_time > last_md_event + IGNORE_TIME )); then
            # Only process if not within mdEvent ignore window
            log "info" "dbEvent detected: $current_time, wait until db updated"
            sleep 30
            log "info" "db synchronization start"
            last_db_event=$current_time
            ./dbSyncExec
            log "info" "db synchronization complete"
        else
            log "info" "dbEvent ignored due to mdEvent"
        fi
    elif [[ "$file" == *.md && "$event" == *"ATTRIB"* ]]; then
        # Handle mdEvent
        if (( current_time > last_db_event + IGNORE_TIME )); then
            # Only process if not within dbEvent ignore window
            log "info" "mdEvent detected: $file at $current_time"
            sleep 20
            ./mdSyncExec "$file"
            last_md_event=$current_time
        else
            log "info" "mdEvent ignored due to dbEvent"
        fi
    fi
done
