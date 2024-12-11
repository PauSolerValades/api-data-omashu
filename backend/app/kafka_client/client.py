from confluent_kafka import Producer, Consumer
from app.core.config import settings
import socket


producer_example = Producer({
    'bootstrap.servers': f"{settings.KAFKA_BROKER}:{settings.KAFKA_PORT}", 
    "client.id": "backend",
})

consumer_matches_ids = Consumer({
    'bootstrap.servers': f"{settings.KAFKA_BROKER}:{settings.KAFKA_PORT}",
    "client.id": "backend",
    "group.id": "super grup"

})