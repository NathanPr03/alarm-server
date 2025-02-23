#!/bin/bash

# API Base URL
API_URL="http://localhost:8080"

# Test Sound File
TEST_SOUND="test_sound.wav"

# Check if the test sound file exists
if [ ! -f "$TEST_SOUND" ]; then
    echo "Error: Test sound file '$TEST_SOUND' not found!"
    exit 1
fi

# ✅ 1. Upload Sound File
echo "Uploading sound file..."
UPLOAD_RESPONSE=$(curl -s -X POST "$API_URL/sound" \
    -H "Content-Type: multipart/form-data" \
    -F "file=@$TEST_SOUND")

UPLOAD_STATUS=$(echo "$UPLOAD_RESPONSE" | jq -r '.message // empty')

if [[ -z "$UPLOAD_STATUS" ]]; then
    echo "❌ Upload failed! Response: $UPLOAD_RESPONSE"
    exit 1
else
    echo "✅ Sound uploaded successfully: $UPLOAD_STATUS"
fi

# ✅ 2. Get List of Available Sounds
echo "Fetching list of available sounds..."
SOUNDS_RESPONSE=$(curl -s -X GET "$API_URL/sounds")

SOUNDS=$(echo "$SOUNDS_RESPONSE" | jq -r '.sounds[]')

if [[ -z "$SOUNDS" ]]; then
    echo "❌ No sounds found! Response: $SOUNDS_RESPONSE"
    exit 1
else
    echo "✅ Available sounds:"
    echo "$SOUNDS"
fi

# ✅ 3. Generate RFC3339 Timestamp
# macOS uses 'gdate' from coreutils, RPI uses 'date'
if command -v gdate &> /dev/null; then
    TIMESTAMP=$(gdate -u -d "+1 minute" +"%Y-%m-%dT%H:%M:%SZ")
else
    TIMESTAMP=$(date -u -d "+1 minute" +"%Y-%m-%dT%H:%M:%SZ")
fi

SELECTED_SOUND=$(echo "$SOUNDS" | head -n 1) # Pick the first sound

echo "Scheduling alarm for $TIMESTAMP with sound '$SELECTED_SOUND'..."
SCHEDULE_RESPONSE=$(curl -s -X POST "$API_URL/schedule" \
    -H "Content-Type: application/json" \
    -d "{\"timestamp\":\"$TIMESTAMP\", \"sound\":\"$SELECTED_SOUND\"}")

SCHEDULE_MESSAGE=$(echo "$SCHEDULE_RESPONSE" | jq -r '.message // empty')

if [[ -z "$SCHEDULE_MESSAGE" ]]; then
    echo "❌ Scheduling failed! Response: $SCHEDULE_RESPONSE"
    exit 1
else
    echo "✅ Alarm scheduled successfully: $SCHEDULE_MESSAGE"
fi

echo "🎉 All tests passed successfully!"