from __future__ import annotations

import hashlib
import hmac

from fastapi import Request, Response

from apps.api.settings import settings


def demo_unlock_enabled() -> bool:
    return bool(settings.demo_real_mode_password)


def is_demo_unlocked(request: Request) -> bool:
    if not demo_unlock_enabled():
        return False
    cookie = request.cookies.get(settings.demo_unlock_cookie_name)
    if not cookie:
        return False
    return hmac.compare_digest(cookie, _expected_cookie_value())


def set_demo_unlock_cookie(response: Response) -> None:
    response.set_cookie(
        key=settings.demo_unlock_cookie_name,
        value=_expected_cookie_value(),
        httponly=True,
        secure=True,
        samesite="none",
        max_age=60 * 60 * 8,
        path="/",
    )


def clear_demo_unlock_cookie(response: Response) -> None:
    response.delete_cookie(
        key=settings.demo_unlock_cookie_name,
        httponly=True,
        secure=True,
        samesite="none",
        path="/",
    )


def validate_demo_password(password: str) -> bool:
    if not demo_unlock_enabled():
        return False
    return hmac.compare_digest(password, settings.demo_real_mode_password)


def _expected_cookie_value() -> str:
    secret = settings.demo_unlock_cookie_secret or settings.demo_real_mode_password
    digest = hmac.new(
        key=secret.encode("utf-8"),
        msg=b"demo-real-mode-unlocked",
        digestmod=hashlib.sha256,
    ).hexdigest()
    return f"v1:{digest}"
