#!/bin/bash
DOCKERFILE=$1
DATA=$2
SCRIPT=$3

if [ -z "$DOCKERFILE" ] || [ -z "$DATA" ] || [ -z "$SCRIPT" ]; then
    echo "Usage: $0 <path_to_dockerfile> <path_to_data> <script_to_run>"
    exit 1
fi

echo "Dockerfile: $DOCKERFILE"
echo "Data: $DATA"
echo "Script to run: $SCRIPT"

shift 3  # Remove the first 3 arguments
SCRIPT_ARGS="$@"  # Capture all remaining arguments

# Dynamically source the environment and execute the script within Docker
docker run --rm \
    --network api-data-network \
    -v "$(pwd)/$DATA":/sample-data \
    -v "$(pwd)/scripts":/scripts \
    -v "$(pwd)/output":/output \
    --entrypoint "/bin/bash" \
    $(docker build -q -f "$DOCKERFILE" .) \
    -c "source /usr/local/bin/.env && /scripts/$(basename "$SCRIPT") $SCRIPT_ARGS"
