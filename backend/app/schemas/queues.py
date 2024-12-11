from pydantic import BaseModel, field_validator
from uuid import UUID


class RegionValidator(BaseModel):
    region: str

    @field_validator("region")
    def validate_region(cls, value):
        allowed_regions = [
            "europe",
            "asia",
            "americas",
            "sea",
            "esports",
        ]
        if value not in allowed_regions:
            raise ValueError(
                f"Invalid region '{value}'. Must be one of {allowed_regions}."
            )
        return value


class QueueValidator(BaseModel):
    queue: str

    @field_validator("queue")
    def validate_name(cls, value):
        allowed_queues = ["riotId", "matchIds", "match"]
        if value not in allowed_queues:
            raise ValueError(
                f"Invalid region '{value}'. Must be one of {allowed_queues}."
            )
        return value


class QueueRequest(BaseModel):
    region: str
    name: str

    @field_validator("region")
    def validate_region(cls, value):
        region_validator = RegionValidator(region=value)
        return region_validator.region

    @field_validator("name")
    def validate_name(cls, value):
        queue_validator = QueueValidator(queue=value)
        return queue_validator.queue


class MatchIdsRequest(BaseModel):
    puuid: UUID
    omashuId: UUID
    startTimestamp: int
    endTimestamp: int

    @field_validator("startTimestamp", "endTimestamp")
    def validate_timestamps(cls, value, field):
        if value < 0:
            raise ValueError(f"{field.name} must be a positive integer.")
        return value
