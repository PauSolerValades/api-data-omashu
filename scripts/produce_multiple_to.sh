#!/bin/bash
if [ $# -ne 3 ]; then
    echo "[PRODUCE_TO.SH] Usage: produce_from.sh <TOPIC_NAME> <SCHEMA_ID> <DATA_FILE>"
    exit 1
fi

TOPIC_NAME=$1
SCHEMA_ID=$2
MESSAGE_FILE=$3

# Load environment variables from .env file if it exists
if [ -f /usr/local/bin/.env ]; then
  set -o allexport
  source /usr/local/bin/.env
  set -o allexport
fi

echo "[·] Connecting to $KAFKA_BROKER:$KAFKA_PORT topic $TOPIC_DOWNLOAD_START"

# Iterate through each line in the MESSAGE_FILE and send it to Kafka
while IFS= read -r line; do
    echo "[·] Producing message: $line"

    echo "$line" | kafka-avro-console-producer \
        --broker-list "${KAFKA_BROKER}:${KAFKA_PORT}" \
        --topic "${TOPIC_NAME}" \
        --property schema.registry.url=http://schema-registry:${SCHEMA_REGISTRY_PORT} \
        --property value.schema.id=${SCHEMA_ID} \
        --property parse.key=false \
        --property avro.use.logical.type.converters=true \
        --property value.converter.use.specific.type.converters=true \
        --property value.converter.use.optional.type.converters=true
        
done < ${MESSAGE_FILE}

#the last three property allows us to use complex types

echo "[+] Finished"

# add the following to produce to a topic with a key
#    --property key.schema.id=6 \
#    --property parse.key=true \
#    --property key.schema='{"type":"string"}' \
#    --property key.separator=":" \
