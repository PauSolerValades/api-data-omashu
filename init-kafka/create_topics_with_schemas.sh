#!/bin/bash

LOG_DIR="/app/logs"
mkdir -p $LOG_DIR
SCHEMA_DIR="/avro-schemas"
SCHEMA_REGISTRY_URL="http://schema-registry:$SCHEMA_REGISTRY_PORT"

TOPICS_CONFIG_FILE="topics_and_schemas.txt"

# Create Kafka topics and register schemas
while read -r TOPIC_CONFIG; do
  IFS=' ' read -ra TOPIC_PARTS <<< "$TOPIC_CONFIG"
  TOPIC_NAME="${TOPIC_PARTS[0]}"
  PARTITIONS="${TOPIC_PARTS[1]}"
  REPLICATION_FACTOR="${TOPIC_PARTS[2]}"
  SCHEMA_FILE="${TOPIC_PARTS[3]}"
  HAS_KEY_SCHEMA="${TOPIC_PARTS[4]}"
  CLEANUP_POLICY="${TOPIC_PARTS[5]}"
  MAKE_IT_TOPIC="${TOPIC_PARTS[6]}"

  echo "[·] Registering schema for topic: $TOPIC_NAME"
  SUBJECT="${TOPIC_NAME}-value"
  # tr -d '\r' removes carriage returns that are common in Windows line endings.
  SCHEMA=$(cat "$SCHEMA_DIR/$SCHEMA_FILE" | sed 's/"/\\"/g' | tr -d '\n' | tr -d '\r')

  # Add references if the schema is for StatsPlayer
  if [ "$TOPIC_NAME" == "stats-player" ]; then
    curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" --data "{\"schema\": \"$SCHEMA\", \"references\": [{\"name\": \"statistics.player.SpiderGraph\", \"subject\": \"stats-player-spider-graph-value\", \"version\": 1}, {\"name\": \"statistics.player.Global\", \"subject\": \"stats-player-global-value\", \"version\": 1}]}" "$SCHEMA_REGISTRY_URL/subjects/$SUBJECT/versions"
  elif [ "$TOPIC_NAME" == "stats-match" ]; then
     curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" --data "{\"schema\": \"$SCHEMA\", \"references\": [{\"name\": \"statistics.match.Summary\", \"subject\": \"stats-match-summary-value\", \"version\": 1}, {\"name\": \"statistics.match.Contribution\", \"subject\": \"stats-match-contribution-value\", \"version\": 1}]}" "$SCHEMA_REGISTRY_URL/subjects/$SUBJECT/versions"
  else
    # Normal schema registration
    curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" --data "{\"schema\": \"$SCHEMA\"}" "$SCHEMA_REGISTRY_URL/subjects/$SUBJECT/versions"
  fi

  echo "[+] Schema registered for $SUBJECT"

  if [ "$HAS_KEY_SCHEMA" == "true" ]; then
    # Again, tr -d '\r' removes Windows carriage returns to ensure compatibility.
    KEY_SCHEMA=$(cat "$SCHEMA_DIR/$(basename "$SCHEMA_FILE" .avsc)-key.avsc" | sed 's/"/\\"/g' | tr -d '\n' | tr -d '\r')
    curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" --data "{\"schema\": \"$KEY_SCHEMA\"}" "$SCHEMA_REGISTRY_URL/subjects/${TOPIC_NAME}-key/versions"
    echo "Key schema registered for $TOPIC_NAME-key"
  fi

  echo "[·] Creating topic: $TOPIC_NAME with partitions: $PARTITIONS and replication factor: $REPLICATION_FACTOR"
  if [ "$MAKE_IT_TOPIC" == 0 ]; then
    echo "[+] File $TOPIC_NAME must not be a topic, skipping"
    continue
  fi
  
  if [ "$CLEANUP_POLICY" == "compact" ]; then
    kafka-topics --create --topic "$TOPIC_NAME" --config cleanup.policy="$CLEANUP_POLICY" --bootstrap-server kafka:${KAFKA_PORT} --partitions "$PARTITIONS" --replication-factor "$REPLICATION_FACTOR"
  else
    kafka-topics --create --if-not-exists --topic "$TOPIC_NAME" --bootstrap-server kafka:${KAFKA_PORT} --partitions "$PARTITIONS" --replication-factor "$REPLICATION_FACTOR"
  fi
  echo "[+] Topic $TOPIC_NAME created!"

  
done < "$TOPICS_CONFIG_FILE"

echo "[+] Kafka topics created and Avro schemas registered successfully!"
