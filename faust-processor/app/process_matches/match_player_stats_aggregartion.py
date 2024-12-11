"""import faust
import logging
from datetime import datetime
from app import app
from app.config import settings

logger = logging.getName(__name__)


class PlayerEvent(faust.Record, serializer="json"):
    timestamp: datetime
    player: str
    position: str
    champion: str
    metric: float


class AggregatedData(faust.Record, serializer="json"):
    player: str
    position: str
    champion: str
    month: str
    average_metric: float


input_topic = app.topic("processed_data", value_type=PlayerEvent)
summarized_topic = app.topic("summarized_data", value_type=AggregatedData)

# Table to hold in-memory aggregates
aggregates_table = app.Table("aggregates_table", default=dict)


def is_current_month(timestamp):
    event_month = timestamp.strftime("%Y-%m")
    current_month = datetime.now().strftime("%Y-%m")
    return event_month == current_month


@app.agent(input_topic)
async def process_data(stream):
    async for event in stream:
        month = event.timestamp.strftime("%Y-%m")
        key = f"{event.player}:{event.position}:{event.champion}:{month}"

        # Update the aggregate
        agg = aggregates_table[key]
        agg["sum"] = agg.get("sum", 0) + event.metric
        agg["count"] = agg.get("count", 0) + 1
        agg["average_metric"] = agg["sum"] / agg["count"]

        aggregated_value = AggregatedData(
            player=event.player,
            position=event.position,
            champion=event.champion,
            month=month,
            average_metric=agg["average_metric"],
        )

        if is_current_month(event.timestamp):
            # Produce to Kafka topic for current month
            await summarized_topic.send(key=key, value=aggregated_value)
        else:
            logger.info("Store in PostgreSQL (NOT IMPLEMENTED)")
            # Store in PostgreSQL for past months
            # await store_in_postgresql(aggregated_value)
"""
