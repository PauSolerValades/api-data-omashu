from typing import Any

from fastapi import APIRouter
from fastapi.responses import JSONResponse

router = APIRouter()


@router.get("/healthcheck")
def health() -> Any:
    
    return JSONResponse(content="Hello world from API-DATA FastAPI backend", status_code=200)
