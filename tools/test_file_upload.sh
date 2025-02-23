#!/bin/bash

# API Base URL
API_URL="http://localhost:8080"

# Test Sound File
TEST_SOUND="test_sound.wav"

# Check if the test sound file exists locally
if [ ! -f "$TEST_SOUND" ]; then
    echo "❌ Error: Test sound file '$TEST_SOUND' not found!"
    exit 1
fi

# ✅ 1. Upload Sound File
echo "Uploading sound file..."
UPLOAD_RESPONSE=$(curl -s -X POST "$API_URL/sound" \
    -H "Content-Type: multipart/form-data" \
    -F "file=@$TEST_SOUND")

UPLOAD_MESSAGE=$(echo "$UPLOAD_RESPONSE" | jq -r '.message // empty')
UPLOAD_FILENAME=$(echo "$UPLOAD_RESPONSE" | jq -r '.filename // empty')

if [[ -z "$UPLOAD_MESSAGE" ]]; then
    echo "❌ Upload failed! Response: $UPLOAD_RESPONSE"
    exit 1
else
    echo "✅ Upload Response: $UPLOAD_MESSAGE"
fi

# ✅ 2. Check If the File Exists in the Sounds Directory
if [ -f "../sounds/$UPLOAD_FILENAME" ]; then
    echo "✅ File '$UPLOAD_FILENAME' exists in /sounds directory."
else
    echo "❌ File '$UPLOAD_FILENAME' not found in /sounds directory."
    exit 1
fi

# ✅ 3. Fetch List of Available Sounds
echo "Fetching list of available sounds..."
SOUNDS_RESPONSE=$(curl -s -X GET "$API_URL/sounds")

# Validate JSON response
if [[ "$(echo "$SOUNDS_RESPONSE" | jq -e '.sounds' 2>/dev/null)" == "null" ]]; then
    echo "❌ Invalid JSON response from API: $SOUNDS_RESPONSE"
    exit 1
fi

SOUNDS=$(echo "$SOUNDS_RESPONSE" | jq -r '.sounds[]')

if [[ -z "$SOUNDS" ]]; then
    echo "❌ No sounds found in the API response!"
    exit 1
else
    echo "✅ Available sounds:"
    echo "$SOUNDS"
fi

# ✅ 4. Generate RFC3339 Timestamp for Alarm
if command -v gdate &> /dev/null; then
    TIMESTAMP=$(gdate -u -d "+1 minute" +"%Y-%m-%dT%H:%M:%SZ")
else
    TIMESTAMP=$(date -u -d "+1 minute" +"%Y-%m-%dT%H:%M:%SZ")
fi

SELECTED_SOUND=$(echo "$SOUNDS" | head -n 1) # Pick the first available sound

if [[ -z "$SELECTED_SOUND" ]]; then
    echo "❌ No valid sound found for scheduling."
    exit 1
fi

# ✅ 5. Schedule an Alarm with the Uploaded Sound
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