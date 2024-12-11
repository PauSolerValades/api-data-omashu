#!/bin/bash
set -e

if [ $# -ne 2 ]; then
    echo "[CONSUME_FROM.SH] Please provide the topic where you want to consume and where you want to store it"
    exit 1
fi

TOPIC=$1
FILENAME=$2

# Load environment variables from .env file if it exists
if [ -f /usr/local/bin/.env ]; then
    set -o allexport
    source /usr/local/bin/.env
    set +o allexport
fi

echo "[DEBUG] KAFKA_BROKER: $KAFKA_BROKER"
echo "[DEBUG] KAFKA_PORT: $KAFKA_PORT"
echo "[DEBUG] SCHEMA_REGISTRY_PORT: $SCHEMA_REGISTRY_PORT"

SCHEMA_REGISTRY_URL="http://schema-registry:$SCHEMA_REGISTRY_PORT/"

echo "[·] Connecting to $KAFKA_BROKER:$KAFKA_PORT"
echo "[·] Using schema '$TOPIC.avsc' from $SCHEMA_REGISTRY_URL"

mkdir -p /output
touch /output/${FILENAME}

kafka-avro-console-consumer \
    --bootstrap-server "${KAFKA_BROKER}:${KAFKA_PORT}" \
    --topic "${TOPIC}" \
    --property schema.registry.url=${SCHEMA_REGISTRY_URL} \
    --from-beginning \
    >> /output/${FILENAME}

echo "[INFO] Consumer finished"
