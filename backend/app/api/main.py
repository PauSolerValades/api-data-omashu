from fastapi import APIRouter

from app.api.routes import healthcheck, queues, matches_id

api_router = APIRouter()
api_router.include_router(healthcheck.router, tags=["healthcheck"])
api_router.include_router(matches_id.router, tags=["matches_id"])
api_router.include_router(queues.router, tags=["queues"])
