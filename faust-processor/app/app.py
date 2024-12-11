import faust

from .config import settings

app = faust.App(
    "match_data_processor",
    version=1,
    autodiscover=["app.process_matches"],
    broker=f"kafka://{settings.KAFKA_BROKER}:{settings.KAFKA_PORT}",
    store="memory://",
    topic_partitions=1,
)

app.conf.producer_max_request_size = 10_485_760  # 10 MB


def main() -> None:
    app.main()
