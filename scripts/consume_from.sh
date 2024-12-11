#!/bin/bash
set -e

if [ $# -ne 1 ]; then
    echo "[CONSUME_FROM.SH] Please provide the topic where you want to consume"
    exit 1
fi

TOPIC=$1

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


kafka-avro-console-consumer \
    --bootstrap-server "${KAFKA_BROKER}:${KAFKA_PORT}" \
    --topic "${TOPIC}" \
    --property schema.registry.url=${SCHEMA_REGISTRY_URL} \
    --from-beginning \
    --max-messages 10 \
    --timeout-ms 10000

echo "[INFO] Consumer finished"