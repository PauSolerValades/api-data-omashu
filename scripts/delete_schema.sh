#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: delete_schema.sh <schema-name>"
    exit 1
fi

SCHEMA_NAME="$1"

# Find the .env file
ENV_FILE=$(find $(pwd) -maxdepth 2 -name '.env' | head -n 1)

if [ -z "$ENV_FILE" ]; then
    echo "Error: .env file not found in the current directory or its parent."
    exit 1
fi

set -a
source "$ENV_FILE"
set +a

echo "Deleting entire schema: ${SCHEMA_NAME}"
RESPONSE=$(docker compose exec -T backend curl -s -X DELETE "http://schema-registry:${SCHEMA_REGISTRY_PORT}/subjects/${SCHEMA_NAME}")
echo "$RESPONSE"

if [[ "$RESPONSE" == *"error_code"* ]]; then
    echo "Failed to delete schema. Please check if the schema exists and try again."
else
    echo "Schema deletion complete."
fi