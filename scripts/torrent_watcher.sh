#!/bin/bash

. /home/efirlus/goproject/scripts/logger.sh
log_file="/home/efirlus/goproject/Logs/app.log"

log "info" "categorize initiated"

# Directories
download_dir="/NAS4/ts"
korean_dir="/NAS4/ts/k"

# Function to check if a file contains Korean characters
contains_korean() {
    local string="$1"
    if echo -n "$string" | grep -P '[가-힣]'; then
        return 0  # true
    fi
    return 1  # false
}

# Function to check if a file should be categorized as Korean
is_korean_file() {
    local filename="$1"
    if contains_korean "$filename" ||
        [[ "$filename" =~ [가-힣] ]] ||
        [[ "$filename" == *韓* ]] ||
        [[ "$filename" == *韩* ]] ||
        [[ "${filename,,}" == *korean* ]]; then
        return 0  # true
    else
        return 1  # false
    fi
}

# Monitor the download directory for new files
inotifywait -m -e close_write --format "%f" "$download_dir" | while IFS= read -r filename; do
    # Full path of the downloaded file
    file_path="$download_dir/$filename"

    # Exclude intermediate files
    if [[ "$filename" == *.crdownload || "$filename" == *.tmp || "$filename" == *.part ]]; then
        continue
    fi

    MIME_TYPE=$(file --mime-type -b "$file_path")

    if [[ "$MIME_TYPE" == application/x-bittorrent ]]; then
        sleep 20
        # Remove common prefixes from the filename (e.g., {torrent - do not redistribute})
        cleaned_filename=$(echo "$filename" | sed 's/{EHT PERSONALIZED TORRENT - DO NOT REDISTRIBUTE}//g' | sed 's/^[[:space:]]*//')

        # Determine where the file should go (Korean directory or remain in the current directory)
        if is_korean_file "$cleaned_filename"; then
            target_dir="$korean_dir"
        else
            target_dir="$download_dir"
        fi

        new_file_path="$target_dir/$cleaned_filename"

        # Check if the file already exists in the target directory
        if [[ -f "$new_file_path" ]]; then
            continue
        fi

        if mv -- "$file_path" "$new_file_path"; then
            if [[ "$target_dir" != "$korean_dir" ]]; then
                log "info" "$filename -> $cleaned_filename file renamed"
            else
                log "info" "$cleaned_filename file moved"
            fi
        fi
    fi
done

log "info" "torrent categorization service end"