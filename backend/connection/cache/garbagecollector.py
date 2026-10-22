"""Client for triggering the Go garbage collector over localhost HTTP."""

from __future__ import annotations

import json
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

DEFAULT_ENDPOINT = "http://127.0.0.1:8765/collect"


def collect(
    endpoint: str = DEFAULT_ENDPOINT, timeout: float = 10.0
) -> dict[str, Any]:
    """Ask the Go collector to scan its configured directory.

    The Go service controls the directory; this client cannot choose a path.
    """
    request = Request(endpoint, data=b"", method="POST")
    try:
        with urlopen(request, timeout=timeout) as response:
            result = json.load(response)
    except HTTPError as error:
        detail = error.read().decode("utf-8", errors="replace")
        raise RuntimeError(
            f"garbage collector returned HTTP {error.code}: {detail}"
        ) from error
    except URLError as error:
        raise RuntimeError(f"could not reach garbage collector: {error}") from error

    if not isinstance(result, dict) or not isinstance(result.get("deleted"), list):
        raise RuntimeError("garbage collector returned an invalid response")
    if result.get("error"):
        raise RuntimeError(
            f"garbage collector reported an error: {result['error']}; "
            f"deleted files: {result['deleted']}"
        )
    return result
