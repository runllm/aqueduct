# generated by datamodel-codegen:
#   filename:  pet_example.json
#   timestamp: 2023-04-04T21:29:57+00:00

from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any, List, Optional

from pydantic import BaseModel, Field


class Model(BaseModel):
    __root__: Any


class ApiResponse(BaseModel):
    code: Optional[int] = None
    type: Optional[str] = None
    message: Optional[str] = None


class Category(BaseModel):
    id: Optional[int] = None
    name: Optional[str] = None


class Status(Enum):
    available = 'available'
    pending = 'pending'
    sold = 'sold'


class Tag(BaseModel):
    id: Optional[int] = None
    name: Optional[str] = None


class Status1(Enum):
    placed = 'placed'
    approved = 'approved'
    delivered = 'delivered'


class Order(BaseModel):
    id: Optional[int] = None
    petId: Optional[int] = None
    quantity: Optional[int] = None
    shipDate: Optional[datetime] = None
    status: Optional[Status1] = Field(None, description='Order Status')
    complete: Optional[bool] = None


class User(BaseModel):
    id: Optional[int] = None
    username: Optional[str] = None
    firstName: Optional[str] = None
    lastName: Optional[str] = None
    email: Optional[str] = None
    password: Optional[str] = None
    phone: Optional[str] = None
    userStatus: Optional[int] = Field(None, description='User Status')


class Pet(BaseModel):
    id: Optional[int] = None
    category: Optional[Category] = None
    name: str = Field(..., example='doggie')
    photoUrls: List[str]
    tags: Optional[List[Tag]] = None
    status: Optional[Status] = Field(None, description='pet status in the store')
