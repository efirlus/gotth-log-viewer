#!/bin/bash

. /home/efirlus/goproject/scripts/logger.sh
log_file="/home/efirlus/goproject/Logs/app.log"

# Configuration
WATCH_DIR_1="/NAS4/MMD"
WATCH_DIR_2="/NAS3/samba/Fancam"
WATCH_DIR_3="/NAS2/priv/PMV"
PATHLOG_DIR="/home/efirlus/goproject/Logs"

# Initialize log files
touch "$PATHLOG_DIR/MMD.pathlog" "$PATHLOG_DIR/Fancam.pathlog" "$PATHLOG_DIR/PMV.pathlog"

# Declare associative array for tracking last logged times
declare -A last_logged

# Function to update pathlog with debouncing
update_pathlog() {
    local dir_num=$1
    local event_path=$2
    
    # Get current time in seconds since epoch
    local current_time=$(date +%s)
    
    # Get last logged time for this file, default to 0 if not set
    local last_time=${last_logged["${event_path}"]-0}
    
    # Only process if more than 5 seconds have passed since last log
    if (( current_time - last_time > 5 )); then
        if file --mime-type "$event_path" | grep -q '^.*: video/'; then
            log "info" "$dir_num accessed: $event_path"
            echo "$event_path" > "$PATHLOG_DIR/${dir_num}.pathlog"
            last_logged["${event_path}"]=$current_time
        fi
    fi
}

# Function to monitor directory
monitor_directory() {
    local dir_path=$1
    local dir_num=$2
    
    # Add error checking for directory existence
    if [ ! -d "$dir_path" ]; then
        log "error" "Directory $dir_path does not exist"
        return 1
	fi
    
    inotifywait -m -r -e close_nowrite --format '%w%f' "$dir_path" 2>/dev/null | while read -r path; do
        # Skip hidden files and directories
        if [[ "$path" == *"/."* ]]; then
            continue
        fi
        
        # Process the event with debouncing
        update_pathlog "$dir_num" "$path"
    done &
}

# Log service start
log "info" "night-time playing watch initiated"

# Start monitoring each directory with error handling
for dir in "$WATCH_DIR_1:MMD" "$WATCH_DIR_2:Fancam" "$WATCH_DIR_3:PMV"; do
    IFS=':' read -r path name <<< "$dir"
    monitor_directory "$path" "$name" || log "error" "Failed to start monitoring $path"
done

# Keep script running and handle signals
trap 'log "info" "Service stopping"; exit 0' SIGTERM SIGINT

# Wait indefinitely
while true; do
    sleep 1
done
