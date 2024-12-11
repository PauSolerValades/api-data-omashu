#!/bin/bash

# Check if the correct number of arguments is provided
if [ $# -ne 2 ]; then
    echo "[PRODUCE_TO.SH] Usage: produce_to.sh <TOPIC_NAME> <DATA_FILE>"
    exit 1
fi

TOPIC_NAME=$1
MESSAGE_FILE=$2

# Load environment variables from .env file if it exists
if [ -f /usr/local/bin/.env ]; then
  set -o allexport
  source /usr/local/bin/.env
  set +o allexport
fi

echo "[·] Connecting to $KAFKA_BROKER:$KAFKA_PORT topic $TOPIC_NAME"

# Construct Schema Registry URL
SCHEMA_REGISTRY_URL="http://${SCHEMA_REGISTRY_HOST}:${SCHEMA_REGISTRY_PORT}"

# Fetch the latest schema ID for the topic's value schema
SCHEMA_ID=$(curl -s "${SCHEMA_REGISTRY_URL}/subjects/${TOPIC_NAME}-value/versions/latest" | jq '.id')

# Check if SCHEMA_ID was retrieved successfully
if [ -z "$SCHEMA_ID" ]; then
    echo "[-] Failed to retrieve schema ID for subject ${TOPIC_NAME}-value"
    exit 1
fi

echo "[·] Using schema ID: $SCHEMA_ID"

# Produce messages to the Kafka topic
kafka-avro-console-producer \
    --broker-list "${KAFKA_BROKER}:${KAFKA_PORT}" \
    --topic "${TOPIC_NAME}" \
    --property schema.registry.url=${SCHEMA_REGISTRY_URL} \
    --property value.schema.id=${SCHEMA_ID} \
    --property parse.key=false \
    --property avro.use.logical.type.converters=true \
    --property value.converter.use.specific.type.converters=true \
    --property value.converter.use.optional.type.converters=true \
    < "${MESSAGE_FILE}"

echo "[+] Finished"

# add the following to produce to a topic with a key
#    --property key.schema.id=6 \
#    --property parse.key=true \
#    --property key.schema='{"type":"string"}' \
#    --property key.separator=":" \
