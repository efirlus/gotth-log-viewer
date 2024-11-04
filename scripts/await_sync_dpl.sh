#!/bin/bash

. /home/efirlus/goproject/scripts/logger.sh
log_file="/home/efirlus/goproject/Logs/app.log"

log "info" "sync with watched result"

# Define arrays for log and dpl paths for each directory
declare -A log_paths=(
    ["MMD"]="/home/efirlus/goproject/Logs/MMD.pathlog"
    ["Fancam"]="/home/efirlus/goproject/Logs/Fancam.pathlog"
    ["PMV"]="/home/efirlus/goproject/Logs/PMV.pathlog"
)

declare -A dpl_paths=(
    ["MMD"]="/NAS4/watch/mmd.dpl"
    ["Fancam"]="/NAS3/samba/watch/fancam.dpl"
    ["PMV"]="/NAS2/priv/watch/pmv.dpl"
)

declare -A temp_dpl_paths=(
    ["MMD"]="/NAS4/watch/mmd.tmp"
    ["Fancam"]="/NAS3/samba/watch/fancam.tmp"
    ["PMV"]="/NAS2/priv/watch/pmv.tmp"
)

declare -A base_dirs=(
    ["MMD"]="MMD"
    ["Fancam"]="Fancam"
    ["PMV"]="PMV"
)

declare -A drive_letters=(
    ["MMD"]="Z"
    ["Fancam"]="E"
    ["PMV"]="Y"
)

# Process each directory
for dir_key in "${!log_paths[@]}"; do
    log "info" "Processing ${dir_key}..."
    
    if [[ -f "${log_paths[$dir_key]}" ]]; then
        # Debug: Check if log file exists and is readable
        # log "info" "Found log file: ${log_paths[$dir_key]}"
        
        # Get the current time and the modification time of log
        current_time=$(date +%s)
        log_modtime=$(stat -c %Y "${log_paths[$dir_key]}")

        # Check if log has been modified/created within the last 6 hours (21600 seconds)
        if [ $((current_time - log_modtime)) -le 21600 ]; then
            log "info" "LOG for ${dir_key} is written within 6 hours"

            # Debug: Check if dpl file exists and is readable
            #if [[ ! -f "${dpl_paths[$dir_key]}" ]]; then
            #    log "error" "DPL file ${dpl_paths[$dir_key]} does not exist"
            #    continue
            #fi
            
            # Debug: Check file contents
            #log "info" "Reading from log file: ${log_paths[$dir_key]}"
            name=$(cat "${log_paths[$dir_key]}" | cut -d'/' -f4-)
            log "info" "Extracted path: [ $name ]"

            # Debug: Check dpl file contents
            #log "info" "Reading from DPL file: ${dpl_paths[$dir_key]}"
            
            # Create directory for temp file if it doesn't exist
            temp_dir=$(dirname "${temp_dpl_paths[$dir_key]}")
            mkdir -p "$temp_dir"

            # Write to temporary file with error checking
            if ! {
                while IFS= read -r line; do
                    if [[ $line == playname=* ]]; then
                        new_line="playname=${drive_letters[$dir_key]}:\\$name"
                        echo "$new_line"
                    else
                        echo "$line"
                    fi
                done < "${dpl_paths[$dir_key]}"
            } > "${temp_dpl_paths[$dir_key]}"; then
                log "error" "Failed to write to temporary file ${temp_dpl_paths[$dir_key]}"
                continue
            fi

            # Debug: Check if temp file was created and has content
            #if [[ -f "${temp_dpl_paths[$dir_key]}" ]]; then
                #log "info" "Temp file created: ${temp_dpl_paths[$dir_key]}"
                #log "info" "Temp file size: $(stat -f%z "${temp_dpl_paths[$dir_key]}")"
            #else
                #log "error" "Temp file was not created: ${temp_dpl_paths[$dir_key]}"
            #fi

            # If the temp file exists and has content, move it
            if [[ -s "${temp_dpl_paths[$dir_key]}" ]]; then
                if mv "${temp_dpl_paths[$dir_key]}" "${dpl_paths[$dir_key]}"; then
                    log "info" "Successfully updated ${dpl_paths[$dir_key]}"
                else
                    log "error" "Failed to move temp file to ${dpl_paths[$dir_key]}"
                fi
            else
                log "error" "Temporary file is empty or invalid for ${dir_key}"
                # Debug: If temp file exists but is empty, show permissions
                #if [[ -f "${temp_dpl_paths[$dir_key]}" ]]; then
                #    log "info" "Temp file permissions: $(ls -l "${temp_dpl_paths[$dir_key]}")"
                #fi
            fi
        else
            log "info" "No New Log within 6 hours for ${dir_key}"
        fi
    else
        log "info" "No log file exists for ${dir_key}"
    fi
done

log "info" "Finishing sync script for all directories"