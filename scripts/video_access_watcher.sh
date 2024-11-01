#!/bin/bash

. /home/efirlus/goproject/scripts/logger.sh
log_file="/home/efirlus/goproject/Logs/app.log"

log "info" "night-time playing watch initiated"

# inotify를 활용해 특정 폴더 안의 모든 동영상파일을 recursive하게 파악해
# 해당 파일의 액세스를 체크해서 값을 단 한 줄 반환하는 스크립트

# 대상 디렉토리
watch_dirs=(
    "/NAS4/MMD"
    "/NAS3/samba/Fancam"
    "/NAS2/priv/PMV"
)

# 로그 생성 위치
declare -A path_logs=(
    ["/NAS4/MMD"]="/home/efirlus/goproject/Logs/mmd-path.log"
    ["/NAS3/samba/Fancam"]="/home/efirlus/goproject/Logs/fancam-path.log"
    ["/NAS2/priv/PMV"]="/home/efirlus/goproject/Logs/pmv-path.log"
)

# Debug: Check if the directory to be watched exists
for dir in "${watch_dirs[@]}"; do
    if [ ! -d "$dir" ]; then
        log "error" "Directory $dir does not exist. Exiting."
        exit 1
    fi
done

declare -A LAST_LOG_TIME  
# Create an associative array to store the last log time for each file

# Prepare inotifywait command for multiple directories
inotify_dirs=()
for dir in "${watch_dirs[@]}"; do
    inotify_dirs+=("$dir")
done

# Main monitoring loop
inotifywait -m -r -e access --format '%w%f' "${inotify_dirs[@]}" | while read FILE
do
    if [ -f "$FILE" ]; then
        # 파일 타입을 먼저 확인 후 스킵할 걸 처리해야 더 메모리 효율이 좋을 것
        MIME_TYPE=$(file --mime-type -b "$FILE")
        if [[ "$MIME_TYPE" == video/* ]]; then
            CURRENT_TIME=$(date +%s)  # Get the current time in seconds

            # Check if the file has been logged before, and calculate time difference
            if [[ -n "${LAST_LOG_TIME[$FILE]}" ]]; then
                TIME_DIFF=$((CURRENT_TIME - LAST_LOG_TIME[$FILE]))
            else
                TIME_DIFF=1201  # Default to a value greater than 1200 if the file hasn't been logged yet
            fi

            # Log the file if more than 20 minutes (1200 seconds) have passed since the last log for this specific file
            if [ "$TIME_DIFF" -gt 1200 ]; then
                # Determine which directory the file belongs to and use corresponding log file
                for dir in "${watch_dirs[@]}"; do
                    if [[ "$FILE" == "$dir"* ]]; then
                        echo "$FILE" > "${path_logs[$dir]}"
                        log "info" "[$dir] [ $FILE ] accessed"
                        break
                    fi
                done
                LAST_LOG_TIME[$FILE]=$CURRENT_TIME  # Update the last log time for this file
            fi
        fi
    fi
done

log "info" "Finishing video access watcher script"
