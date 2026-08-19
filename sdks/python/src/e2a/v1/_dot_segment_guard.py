"""Reject a dot-segment path parameter before httpx can resolve it as "go up
one level" and retarget the request onto a parent resource. Called from
``_TypedApiClient.param_serialize`` in ``client.py``, on the raw
``path_params`` dict, before the path string is built."""

from __future__ import annotations

import re
from typing import Mapping, Optional

from .errors import E2AValidationError

_ALL_DOTS = re.compile(r"^\.+$")


def reject_dot_segment_path_params(path_params: Optional[Mapping[str, object]]) -> None:
    if not path_params:
        return
    for name, value in path_params.items():
        if isinstance(value, str) and _ALL_DOTS.match(value):
            raise E2AValidationError(
                code="invalid_request_path",
                message=(
                    f'refusing to send a request whose {name!r} path parameter '
                    f'is a dot-segment ("{value}")'
                ),
                status=0,
                retryable=False,
            )
