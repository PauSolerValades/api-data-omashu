from fastapi import APIRouter, Depends, Path
from fastapi.exceptions import HTTPException
from redis import Redis
from uuid import UUID

from app.core.dependencies import get_redis_client
from app.schemas.queues import QueueRequest, MatchIdsRequest, RegionValidator

router = APIRouter()


def get_queue_request(
    region_id: str = Path(..., description="Identifier of the region"),
    queue_id: str = Path(..., description="Identifier of the queue"),
) -> QueueRequest:
    """
    Dependency function to parse and validate region_id and queue_id using QueueRequest schema.
    """
    return QueueRequest(region_id=region_id, queue_id=queue_id)


def get_match_ids_request(
    puuid: UUID = Path(..., description="Player uuid identifier of LoL"),
    omashuId: UUID = Path(
        ..., description="UUID of the registered user in the frontend."
    ),
    startTimestamp: int = Path(..., description="Lower bound for matches to search."),
    endTimestamp: int = Path(..., descrition="Upper bound for matches to search."),
) -> MatchIdsRequest:
    return MatchIdsRequest(puuid, omashuId, startTimestamp, endTimestamp)


def get_region_request(
    region: str = Path(..., description="Valid regions for riotId"),
) -> RegionValidator:
    """
    Dependency function to parse and validate the region
    """
    return RegionValidator(region)


@router.get(
    "/queue/{region_id}/{queue_id}/size",
    description="Retrieve the number of elements in the specified queue.",
)
def get_queue_length(
    queue_request: QueueRequest = Depends(get_queue_request),
    redis_client: Redis = Depends(get_redis_client),
):
    """
    Endpoint to get the length of a queue.
    """
    queue_name = f"{queue_request.region_id}-{queue_request.queue_id}"
    length = redis_client.llen(queue_name)
    if length is None:
        raise HTTPException(status_code=404, detail="Queue not found")
    return {"queue": queue_name, "length": length}


@router.post(
    "/queue/{region_id}/matchIds/element",
    description="Append an element at the beggining of the queue.",
)
def append_to_queue(
    region_id: RegionValidator = Depends(get_region_request),
    element_request: MatchIdsRequest = Depends(get_match_ids_request),
    redis_client: Redis = Depends(get_redis_client),
):
    queue_name = f"{region_id}-matchIds"
    element_data = element_request.model_dump_json()
    redis_client.lpush(queue_name, element_data)
    return {"message": f"Element added to {queue_name} with priority."}
