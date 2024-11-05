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

# Function to update pathlog
update_pathlog() {
    local dir_num=$1
    local event_path=$2
    echo "$event_path" > "$PATHLOG_DIR/${dir_num}.pathlog"
}

# Function to monitor directory
monitor_directory() {
    local dir_path=$1
    local dir_num=$2
    
	inotifywait -m -r -e close_nowrite --format '%w%f' "$dir_path" | while read -r path; do
        # Check if the file is of video mimetype
        if file --mime-type "$path" | grep -q '^.*: video/'; then
            log "info" "$dir_num accessed: $path"
            update_pathlog "$dir_num" "$path"
        fi
    done &
}

# Log service start
log "info" "night-time playing watch initiated"

# Start monitoring each directory
monitor_directory "$WATCH_DIR_1" "MMD"
monitor_directory "$WATCH_DIR_2" "Fancam"
monitor_directory "$WATCH_DIR_3" "PMV"

# Keep script running and handle signals
trap 'log "info" "Service stopping"; exit 0' SIGTERM SIGINT

# Wait indefinitely
while true; do
    sleep 1
done
