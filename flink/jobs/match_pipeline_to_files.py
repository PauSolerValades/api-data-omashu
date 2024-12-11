import logging
import sys
import json
from pyflink.common import Types

from pyflink.datastream import StreamExecutionEnvironment, KeyedProcessFunction
from pyflink.table import StreamTableEnvironment
from pyflink.datastream.state import ValueStateDescriptor
from pyflink.datastream.connectors.file_system import RollingPolicy
from pydantic_settings import BaseSettings, SettingsConfigDict
from pyflink.datastream.connectors.file_system import (
    FileSink,
    OutputFileConfig,
)
from pyflink.common.serialization import Encoder


import reduce_match_data
from new_metrics import new_metrics


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env", env_ignore_empty=True, extra="ignore"
    )
    PROJECT_NAME: str = "Flink"
    KAFKA_PORT: int
    SCHEMA_REGISTRY_PORT: str
    TOPIC_MATCH_DATA: str


settings = Settings()


def process_json_data(data):
    try:
        json_data = json.loads(data)
        return json_data
    except json.JSONDecodeError:
        return {"error": "Invalid JSON"}


class MatchDataProcessor(KeyedProcessFunction):
    def __init__(self):
        self.postmatch_state = None
        self.bytime_state = None

    def open(self, runtime_context):
        self.postmatch_state = runtime_context.get_state(
            ValueStateDescriptor("POSTMATCH", Types.STRING())
        )
        self.bytime_state = runtime_context.get_state(
            ValueStateDescriptor("BYTIME", Types.STRING())
        )

    def process_element(self, value, ctx: KeyedProcessFunction.Context):
        matchId, obtainedAt, omashuId, puuid, dataType, data = value
        print(
            f"Received data: matchId={matchId}, obtainedAt={obtainedAt}, omashuId={omashuId}, puuid={puuid}, dataType={dataType}"
        )
        print(f"Data is None: {data is None}")

        if matchId is None:
            print("WARNING: Received None matchId. Skipping processing.")
            return

        processed_data = process_json_data(data)
        print(f"Processed data is error: {'error' in processed_data}")

        if dataType == "POSTMATCH":
            print(f"RECEIVED {matchId} POSTMATCH DATA")
            reduced_pm = reduce_match_data.reduce_postmatch_data(processed_data)
            self.postmatch_state.update(json.dumps(reduced_pm))
            print(f"Stored POSTMATCH data for matchId {matchId}")
        elif dataType == "BYTIME":
            print(f"RECEIVED {matchId} BYTIME DATA")
            reduced_bt = reduce_match_data.reduce_bytime_data(processed_data)
            self.bytime_state.update(json.dumps(reduced_bt))
            print(f"Stored BYTIME data for matchId {matchId}")
        else:
            print(f"WARNING: Unknown dataType {dataType}")
            return

        postmatch_data = self.postmatch_state.value()
        bytime_data = self.bytime_state.value()

        print(
            f"State after update: postmatch exists={postmatch_data is not None}, bytime exists={bytime_data is not None}"
        )

        if postmatch_data and bytime_data:
            print(
                f"Both POSTMATCH and BYTIME data available for matchId {matchId}. Processing..."
            )
            try:
                postmatch_json = json.loads(postmatch_data)
                bytime_json = json.loads(bytime_data)

                new_metrics(bytime_json, postmatch_json)
                # Create a readable output
                output = {
                    "matchId": matchId,
                    "POSTMATCH": postmatch_json,
                    "BYTIME": bytime_json,
                }

                self.postmatch_state.clear()
                self.bytime_state.clear()

                print(f"Processing complete for matchId: {matchId}")
                super_output = json.dumps(output)
                print(super_output)
                print(type(super_output))
                yield super_output
            except json.JSONDecodeError as e:
                print(f"Error decoding JSON for matchId {matchId}: {e}")
            except Exception as e:
                print(f"Error processing data for matchId {matchId}: {e}")


def kafka_confluent_avro_job():
    env = StreamExecutionEnvironment.get_execution_environment()
    t_env = StreamTableEnvironment.create(stream_execution_environment=env)

    t_env.execute_sql(f"""
        CREATE TABLE kafka_source (
            matchId STRING,
            obtainedAt INT,
            omashuId STRING,
            puuid STRING,
            dataType STRING,
            data STRING
        ) WITH (
            'connector' = 'kafka',
            'topic' = '{settings.TOPIC_MATCH_DATA}',
            'properties.bootstrap.servers' = 'kafka:{settings.KAFKA_PORT}',
            'properties.group.id' = 'flink-consumer-group',
            'scan.startup.mode' = 'earliest-offset',
            'format' = 'avro-confluent',
            'avro-confluent.schema-registry.url' = 'http://schema-registry:{settings.SCHEMA_REGISTRY_PORT}'
        )
    """)

    source_table = t_env.from_path("kafka_source")
    ds = t_env.to_data_stream(source_table)

    mdp = MatchDataProcessor()
    ds = (
        ds.key_by(lambda row: row.matchId)
        .process(mdp, output_type=Types.STRING())
        .filter(lambda x: x is not None)
    )

    output_path = "/opt/flink/output"
    ds.sink_to(
        sink=FileSink.for_row_format(
            base_path=output_path, encoder=Encoder.simple_string_encoder()
        )
        .with_output_file_config(OutputFileConfig.builder().build())
        .with_rolling_policy(RollingPolicy.default_rolling_policy())
        .build()
    )

    env.execute("Kafka Confluent Avro State Management Job")


if __name__ == "__main__":
    logging.basicConfig(stream=sys.stdout, level=logging.INFO, format="%(message)s")
    kafka_confluent_avro_job()
