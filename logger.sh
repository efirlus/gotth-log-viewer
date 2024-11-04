#!/bin/bash

LOG_FILE="/home/efirlus/goproject/Logs/apptest.log"

log() {
    local level="$1"
    local message="$2"
    local program_name="$(basename "$0")"
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S.%3N")
    local json_log

    json_log=$(jq -nc \
        --arg time "$timestamp" \
        --arg loglevel "$level" \
        --arg programname "$program_name" \
        --arg message "$message" \
        '{time: $time, loglevel: $loglevel, programname: $programname, message: $message}')

    if [[ "$level" =~ ^(error|fatal|panic)$ ]]; then
        local line_number="${BASH_LINENO[0]}"
        local script_name="$(basename "${BASH_SOURCE[1]}")"
        json_log=$(jq -c --arg location "${script_name}:${line_number}" '. + {location: $location}' <<< "$json_log")
    fi

    echo "$json_log"

    if [[ -n "$LOG_FILE" ]]; then
        echo "$json_log" >> "$LOG_FILE"
    fi
}
