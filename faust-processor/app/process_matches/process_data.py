import logging
import json

from schema_registry.client import SchemaRegistryClient
from schema_registry.serializers.faust import FaustSerializer

from app import app
from app.config import settings

import reduce_match_data
from new_metrics import new_metrics
import extract_data

logger = logging.getLogger(__name__)


# Initialize Schema Registry Client
client = SchemaRegistryClient(url=settings.SCHEMA_REGISTRY_URL)


def get_schema_or_fail(client: SchemaRegistryClient, subject_name: str):
    # Try to fetch the schema from the registry
    schema_response = client.get_schema(subject_name)
    if schema_response is None:
        logger.error(f"Schema for {subject_name} not found in Schema Registry.")
        raise Exception(f"Schema for {subject_name} not found in Schema Registry.")
    logger.info(f"Obtained {subject_name} from schema registry")
    return schema_response.schema


try:
    avro_match_data_schema = get_schema_or_fail(
        client, f"{settings.TOPIC_MATCH_DATA}-value"
    )

    avro_match_to_store = get_schema_or_fail(
        client, f"{settings.TOPIC_MATCH_TO_STORE}-value"
    )

    avro_match_player_data_schema = get_schema_or_fail(
        client, f"{settings.TOPIC_MATCH_PLAYER_DATA}-value"
    )

    avro_match_team_data_schema = get_schema_or_fail(
        client, f"{settings.TOPIC_MATCH_TEAM_DATA}-value"
    )

    avro_match_game_data_schema = get_schema_or_fail(
        client, f"{settings.TOPIC_MATCH_GAME_DATA}-value"
    )

except Exception as e:
    logger.error(f"Error fetching schemas: {e}")
    raise

avro_match_event_serializer = FaustSerializer(
    schema_registry_client=client,
    schema_subject=f"{settings.TOPIC_MATCH_DATA}-value",
    schema=avro_match_data_schema,
)

avro_match_to_store_serializer = FaustSerializer(
    schema_registry_client=client,
    schema_subject=f"{settings.TOPIC_MATCH_TO_STORE}-value",
    schema=avro_match_to_store,
)

avro_match_player_data_serializer = FaustSerializer(
    schema_registry_client=client,
    schema_subject=f"{settings.TOPIC_MATCH_PLAYER_DATA}-value",
    schema=avro_match_player_data_schema,
)

avro_match_team_data_serializer = FaustSerializer(
    schema_registry_client=client,
    schema_subject=f"{settings.TOPIC_MATCH_TEAM_DATA}-value",
    schema=avro_match_team_data_schema,
)

avro_match_game_data_serializer = FaustSerializer(
    schema_registry_client=client,
    schema_subject=f"{settings.TOPIC_MATCH_GAME_DATA}-value",
    schema=avro_match_game_data_schema,
)

match_data_topic = app.topic(
    settings.TOPIC_MATCH_DATA,
    key_type=None,
    value_type=None,
    value_serializer=avro_match_event_serializer,
    partitions=1,
)

match_to_store_topic = app.topic(
    settings.TOPIC_MATCH_TO_STORE,
    key_type=None,
    value_type=None,
    value_serializer=avro_match_to_store_serializer,
    partitions=1,
)

match_player_data_topic = app.topic(
    settings.TOPIC_MATCH_PLAYER_DATA,
    key_type=None,
    value_type=None,
    value_serializer=avro_match_player_data_serializer,
    partitions=1,
)

match_team_data_topic = app.topic(
    settings.TOPIC_MATCH_TEAM_DATA,
    key_type=None,
    value_type=None,
    value_serializer=avro_match_team_data_serializer,
    partitions=1,
)

match_game_data_topic = app.topic(
    settings.TOPIC_MATCH_GAME_DATA,
    key_type=None,
    value_type=None,
    value_serializer=avro_match_game_data_serializer,
    partitions=1,
)

match_table = app.Table(settings.TOPIC_MATCH_DATA, default=dict, partitions=1)


@app.agent(match_data_topic)
async def process(stream):
    async for event in stream:
        match_id = event["matchId"]
        data_type = event["dataType"]
        data = json.loads(event["data"])

        if match_id is None:
            continue

        logger.info(f"Received {data_type} {match_id}")

        # Initialize state for a given matchId to track which info has arrived
        if match_id not in match_table:
            match_table[match_id] = {"POSTMATCH": None, "BYTIME": None}

        # Update state
        if data_type == "POSTMATCH":
            reduce_match_data.reduce_postmatch_data(data)
            match_table[match_id]["POSTMATCH"] = data
        elif data_type == "BYTIME":
            reduce_match_data.reduce_bytime_data(data)
            match_table[match_id]["BYTIME"] = data

        logger.info(f"[{data_type} - {match_id}]: Reduced Successfully")

        # Check if both data types are available
        if (
            match_table[match_id]["POSTMATCH"] is not None
            and match_table[match_id]["BYTIME"] is not None
        ):
            logger.info("Starting data transformation")
            reduced_postmatch = match_table[match_id]["POSTMATCH"]
            reduced_bytime = match_table[match_id]["BYTIME"]
            new_metrics(reduced_bytime, reduced_postmatch)

            to_store = {
                "matchId": match_id,
                "POSTMATCH": json.dumps(reduced_postmatch).encode("utf-16"),
                "BYTIME": json.dumps(reduced_bytime).encode("utf-16"),
            }
            logger.info(f"[{match_id}] new_metrics done!")

            # Save original data to S3 by storing into a Kafka topic to be consumed
            await match_to_store_topic.send(value=to_store)
            logger.info(
                f"[{match_id}] Produced to topic {settings.TOPIC_MATCH_TO_STORE}"
            )

            summarized_match = extract_data.compute_stats(
                reduced_postmatch, reduced_bytime
            )
            logger.info(f"[{match_id}] Data extracted! Freeing match from state")
            # Delete the match from state because it already has been processed and stored
            del match_table[match_id]

            # Merge metadata and different stats into a single dict
            metadata: dict = summarized_match["metadata"]
            player_match_data: dict = summarized_match["player_stats"]
            team_match_data: dict = summarized_match["team_stats"]
            game_match_data: dict = summarized_match["game_stats"]

            # Send to corresponding topics
            await match_player_data_topic.send(
                value={
                    "matchId": match_id,
                    "data": json.dumps(
                        {"metadata": metadata, "player": player_match_data}, indent=0
                    ),
                }
            )
            await match_team_data_topic.send(
                value={
                    "matchId": match_id,
                    "data": json.dumps(
                        {"metadata": metadata, "team": team_match_data}, indent=0
                    ),
                }
            )
            await match_game_data_topic.send(
                value={
                    "matchId": match_id,
                    "data": json.dumps(
                        {"metadata": metadata, "team": game_match_data}, indent=0
                    ),
                }
            )
            logger.info(f"[{match_id}] Extracted data sent to match-<type>-data")
