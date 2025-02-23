#!/bin/bash

if [ -z "$1" ]; then
    echo "Error: File name not provider. Usage: $0 <sound_file>"
    exit 1
fi

SOUND_FILE="$1"

if [ ! -f "$SOUND_FILE" ]; then
    echo "Error: File '$SOUND_FILE' not found!"
    exit 1
fi

play "$SOUND_FILE"