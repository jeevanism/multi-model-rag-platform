from __future__ import annotations

from pydantic import BaseModel


class DemoUnlockRequest(BaseModel):
    password: str


class DemoUnlockStatusResponse(BaseModel):
    unlocked: bool
    unlock_enabled: bool
