from fastapi import FastAPI
from fastapi.routing import APIRoute
import asyncio

from app.api.main import api_router
from app.core.config import settings
from app.kafka_client.consumers import consume_loop
from app.kafka_client.client import consumer_matches_ids


def custom_generate_unique_id(route: APIRoute) -> str:
    return f"{route.tags[0]}-{route.name}"


app = FastAPI(
    title=settings.PROJECT_NAME,
    openapi_url=f"{settings.API_V1_STR}/openapi.json",
    generate_unique_id_function=custom_generate_unique_id,
)

app.include_router(api_router, prefix=settings.API_V1_STR)

"""# start up background loops for the consumers
background_loop = asyncio.get_event_loop()
background_loop.create_task(
    consume_loop(consumer_matches_ids, [settings.TOPIC_MATCH_DATA])
)
"""
