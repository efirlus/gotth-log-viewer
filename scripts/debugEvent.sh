#!/bin/bash

# Directory to monitor
MONITOR_DIR="/NAS4/ts"

# Log file
LOG_FILE="/home/efirlus/goproject/Logs/app.log"

# Start monitoring with inotifywait and log all events
inotifywait -m -r \
    -e access -e modify -e attrib -e close_write -e close_nowrite \
    -e open -e moved_to -e moved_from -e create -e delete -e delete_self \
    -e move_self -e unmount \
    "$MONITOR_DIR" |
while read path action file; do
    echo "$(date '+%Y-%m-%d %H:%M:%S') - Path: $path - Action: $action - File: $file" >> "$LOG_FILE"
done