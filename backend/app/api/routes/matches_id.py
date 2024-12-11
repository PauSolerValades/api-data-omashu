from fastapi import FastAPI, APIRouter, HTTPException, Depends, Body
from pydantic import BaseModel, Field
from typing import Any, Generator, Optional, List
from sqlalchemy import create_engine, Column, String, Boolean, ForeignKey
from sqlalchemy.orm import sessionmaker, Session, declarative_base, relationship
from app.models import Base, MatchStatus, Player, PlayerMatch, UpdateMatchStatusModel, MatchIdModel
from app.core.dependencies import get_db

router = APIRouter()

@router.get("/match_id/{matchId}")
def read_match(matchId: str, db: Session = Depends(get_db)) -> Any:
    try:
        MatchIdModel(matchId=matchId)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid matchId format")

    match = db.query(MatchStatus).filter(MatchStatus.match_id == matchId).first()
    if not match:
        raise HTTPException(status_code=404, detail="Match not found")

    return {"matchId": matchId, "status": match.status}

@router.post("/match_id/new")
def create_match(data: dict = Body(...), db: Session = Depends(get_db)) -> Any:
    matchId = data.get("matchId")
    status = data.get("status")

    try:
        MatchIdModel(matchId=matchId)
    except (ValueError, TypeError):
        raise HTTPException(status_code=400, detail="Invalid matchId format")

    if not status or not isinstance(status, str):
        raise HTTPException(status_code=400, detail="Invalid status")

    existing_match = db.query(MatchStatus).filter(MatchStatus.match_id == matchId).first()
    if existing_match:
        raise HTTPException(status_code=400, detail="Match already exists")

    new_match = MatchStatus(match_id=matchId, status=status)

    db.add(new_match)
    db.commit()
    db.refresh(new_match)

    return {"matchId": new_match.match_id, "status": new_match.status}

@router.delete("/match_id/{matchId}")
def delete_match(matchId: str, db: Session = Depends(get_db)) -> dict:
    try:
        MatchIdModel(matchId=matchId)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid matchId format")

    match = db.query(MatchStatus).filter(MatchStatus.match_id == matchId).first()
    if not match:
        raise HTTPException(status_code=404, detail="Match not found")
    
    db.delete(match)
    db.commit()

    return {"matchId deleted": matchId}

@router.get("/match_id/{matchId}/players")
def read_players(matchId: str, player: Optional[str] = None, db: Session = Depends(get_db)) -> Any:
    try:
        MatchIdModel(matchId=matchId)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid matchId format")

    match = db.query(MatchStatus).filter(MatchStatus.match_id == matchId).first()
    if not match:
        raise HTTPException(status_code=404, detail="Match not found")

    players = match.players
    puuids = [player_match.puuid for player_match in players]

    return {"matchId": matchId, "puuids": puuids}

@router.put("/match_id/{matchId}/status")
def update_match_status(matchId: str, update_data: UpdateMatchStatusModel, db: Session = Depends(get_db)) -> Any:
    try:
        MatchIdModel(matchId=matchId)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid matchId format")

    match = db.query(MatchStatus).filter(MatchStatus.match_id == matchId).first()
    if not match:
        raise HTTPException(status_code=404, detail="Match not found")

    current_status = match.status

    if current_status == "QUEUED":
        match.status = "DOWNLOADED"
        db.commit()
        db.refresh(match)
        return {"matchId": matchId, "status": match.status, "message": "Match status updated from QUEUED to DOWNLOADED."}
    else:
        return {"matchId": matchId, "status": match.status, "message": "Match is already in DOWNLOADED status."}
