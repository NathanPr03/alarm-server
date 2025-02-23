#!/bin/bash

LOG_FILE="alarm_sh.log"

log() {
    echo "$(date +"%Y-%m-%d %H:%M:%S") $1" | tee -a "$LOG_FILE"
}

if [ -z "$1" ]; then
    log "❌ Error: File name not provided. Usage: $0 <sound_file>"
    exit 1
fi

SOUND_FILE="$1"

if [ ! -f "$SOUND_FILE" ]; then
    log "❌ Error: File '$SOUND_FILE' not found!"
    exit 1
fi

log "▶️ Playing sound: $SOUND_FILE"

play "$SOUND_FILE" >> "$LOG_FILE" 2>&1

log "✅ Finished playing sound: $SOUND_FILE"