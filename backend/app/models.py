import uuid

from pydantic import EmailStr, BaseModel, Field
from typing import Any, Generator, Optional, List
from sqlmodel import Relationship, SQLModel
from sqlalchemy import create_engine, Column, String, Boolean, ForeignKey
from sqlalchemy.orm import sessionmaker, Session, declarative_base, relationship

# SQLAlchemy setup for matches and players
Base = declarative_base()

class MatchStatus(Base):
    __tablename__ = "match_status"
    
    match_id = Column(String(255), primary_key=True)
    status = Column(String(255))
    players = relationship("PlayerMatch", back_populates="match")

class Player(Base):
    __tablename__ = "player"
    
    puuid = Column(String(255), primary_key=True)
    omashu_id = Column(String(255))
    riot_id = Column(String(255))
    registered = Column(Boolean)
    matches = relationship("PlayerMatch", back_populates="player")

class PlayerMatch(Base):
    __tablename__ = "player_match"
    
    match_id = Column(String(255), ForeignKey("match_status.match_id"), primary_key=True)
    puuid = Column(String(255), ForeignKey("player.puuid"), primary_key=True)
    match = relationship("MatchStatus", back_populates="players")
    player = relationship("Player", back_populates="matches")

# SQLModel for user-related models

class UpdateMatchStatusModel(BaseModel):
    status: str

class MatchIdModel(BaseModel):
    matchId: str = Field(
        ...,
        pattern=r"^(EUW1|EUN1|NA1|BR1|KR|JP1|LA1|LA2|OC1|PH2|RU|SG2|TH2|TW2|TR1|VN2)_[0-9]{10}$"
    )

class UserBase(SQLModel):
    email: EmailStr = Field(unique=True, index=True, max_length=255)
    is_active: bool = True
    is_superuser: bool = False
    full_name: str | None = Field(default=None, max_length=255)

class UserCreate(UserBase):
    password: str = Field(min_length=8, max_length=40)

class UserRegister(SQLModel):
    email: EmailStr = Field(max_length=255)
    password: str = Field(min_length=8, max_length=40)
    full_name: str | None = Field(default=None, max_length=255)

class UserUpdate(UserBase):
    email: EmailStr | None = Field(default=None, max_length=255)  # type: ignore
    password: str | None = Field(default=None, min_length=8, max_length=40)

class UserUpdateMe(SQLModel):
    full_name: str | None = Field(default=None, max_length=255)
    email: EmailStr | None = Field(default=None, max_length=255)

class UpdatePassword(SQLModel):
    current_password: str = Field(min_length=8, max_length=40)
    new_password: str = Field(min_length=8, max_length=40)

class UserPublic(UserBase):
    id: uuid.UUID

class UsersPublic(SQLModel):
    data: list[UserPublic]
    count: int

class ItemBase(SQLModel):
    title: str = Field(min_length=1, max_length=255)
    description: str | None = Field(default=None, max_length=255)

class ItemCreate(ItemBase):
    pass

class ItemUpdate(ItemBase):
    title: str | None = Field(default=None, min_length=1, max_length=255)  # type: ignore

class ItemPublic(ItemBase):
    id: uuid.UUID
    owner_id: uuid.UUID

class ItemsPublic(SQLModel):
    data: list[ItemPublic]
    count: int

class Message(SQLModel):
    message: str

class Token(SQLModel):
    access_token: str
    token_type: str = "bearer"

class TokenPayload(SQLModel):
    sub: str | None = None

class NewPassword(SQLModel):
    token: str
    new_password: str = Field(min_length=8, max_length=40)
