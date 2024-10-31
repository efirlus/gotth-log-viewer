#!/bin/bash

. /home/efirlus/goproject/gobooknotes/logger.sh
LOG_FILE="/home/efirlus/goproject/Logs/app.log"

log "info" "sync initiated"

# Configuration
DB_FILE="/NAS/samba/Book/metadata.db"
ONEDRIVE_DIR="/home/efirlus/OneDrive/obsidian/Vault/6. Calibre"
DEBOUNCE_DELAY=60
DB_DEBOUNCE_DELAY=10
BURST_WINDOW=2
LOCK_FILE="/home/efirlus/goproject/gobooknotes/tmp/watch_lock"
DB_LAST_TRIGGER="/home/efirlus/goproject/gobooknotes/tmp/db_last_trigger"
ONEDRIVE_LAST_TRIGGER="/home/efirlus/goproject/gobooknotes/tmp/onedrive_last_trigger"
DB_EVENT_COUNT="/home/efirlus/goproject/gobooknotes/tmp/db_event_count"

# Ensure required tools are available
if ! command -v inotifywait >/dev/null 2>&1; then
    log "error" "inotifywait not found. Please install inotify-tools package."
    exit 1
fi

# Initialize lock and trigger files
touch "$LOCK_FILE" "$DB_LAST_TRIGGER" "$ONEDRIVE_LAST_TRIGGER"
echo "0" > "$DB_EVENT_COUNT"

# Cleanup on script exit
cleanup() {
    rm -f "$LOCK_FILE" "$DB_LAST_TRIGGER" "$ONEDRIVE_LAST_TRIGGER", "$DB_EVENT_COUNT"
    kill $(jobs -p) 2>/dev/null
    exit 0
}
trap cleanup EXIT INT TERM

# Check if an event should be ignored based on recent triggers
should_ignore_event() {
    local trigger_file=$1
    local current_time=$(date +%s)
    local last_trigger_time=$(cat "$trigger_file")
    
    # Convert empty string to 0
    last_trigger_time=${last_trigger_time:-0}
    
    # If less than DEBOUNCE_DELAY seconds have passed since last trigger
    if (( current_time - last_trigger_time < DEBOUNCE_DELAY )); then
        return 0 # true, should ignore
    fi
    return 1 # false, should not ignore
}

# Record trigger time
record_trigger() {
    local trigger_file=$1
    date +%s > "$trigger_file"
}

# Watch DB file
watch_db() {
    local last_event=0
    local debounce_timer=""
    local is_in_burst=0
    
    while true; do
        inotifywait -e modify "$DB_FILE" 2>/dev/null | while read -r directory events filename; do
            # Skip if event was triggered by onedrive-exec
            if pgrep -f "onedrive-exec" >/dev/null; then
                continue
            fi
            
            # Check if we should ignore this event
            if should_ignore_event "$ONEDRIVE_LAST_TRIGGER"; then
                continue
            fi
            
            current_time=$(date +%s)

            # Increment event count
            count=$(( $(cat "$DB_EVENT_COUNT") + 1 ))
            echo "$count" > "$DB_EVENT_COUNT"
            
            # Clear existing timer if it exists
            if [ ! -z "$debounce_timer" ]; then
                kill $debounce_timer 2>/dev/null
            fi
            
            # Start new timer
            {
                # First wait for a short period to see if more events come
                sleep $DB_DEBOUNCE_DELAY

                # check event count
                final_count=$(cat "$DB_EVENT_COUNT")

                if (( final_count > count )); then
                    # it means it's burst, wait for full debounce period
                    sleep $((DEBOUNCE_DELAY - DB_DEBOUNCE_DELAY))
                fi

                # Reset event count
                echo "0" > "$DB_EVENT_COUNT"

                # Only execute if burst is over
                if ! should_ignore_event "$DB_LAST_TRIGGER"; then
                    record_trigger "$DB_LAST_TRIGGER"
                    log "info" "Executing db-exec command..."
                    /home/efirlus/goproject/gobooknotes/dbSyncExec
                fi
            } &
            debounce_timer=$!
        done
    done
}

# Watch onedrive directory
watch_onedrive() {
    local last_event=0
    local debounce_timer=""
    local last_file=""
    
    while true; do
        inotifywait -r -e attrib --format '%w%f' "$ONEDRIVE_DIR" 2>/dev/null | while read -r fullpath; do
            # Skip if event was triggered by db-exec
            if pgrep -f "db-exec" >/dev/null; then
                continue
            fi
            
            # Check if we should ignore this event
            if should_ignore_event "$DB_LAST_TRIGGER"; then
                continue
            fi
            
            current_time=$(date +%s)

            last_file="$fullpath"

            # Clear existing timer if it exists
            if [ ! -z "$debounce_timer" ]; then
                kill $debounce_timer 2>/dev/null
            fi
            
            # Start new timer
            {
                sleep $DEBOUNCE_DELAY
                record_trigger "$ONEDRIVE_LAST_TRIGGER"
                log "info" "Executing onedrive-exec command for $last_file"
                printf -v quoted_path %q "$last_file"
                eval `/home/efirlus/goproject/gobooknotes/mdSyncExec "$last_file"`
            } &
            debounce_timer=$!
        done
    done
}

# Main execution
main() {
    # Check if files/directories exist
    if [ ! -f "$DB_FILE" ]; then
        log "error" "$DB_FILE does not exist"
        exit 1
    fi
    
    if [ ! -d "$ONEDRIVE_DIR" ]; then
        log "error" "$ONEDRIVE_DIR directory does not exist"
        exit 1
    fi
    
    # Start watchers in background
    watch_db &
    watch_onedrive &
    
    # Wait for both watchers
    wait
}

main