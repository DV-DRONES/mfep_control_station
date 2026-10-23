"""TCP transfer helpers for JSON and YAML files over a Wi-Fi/LAN connection."""

from __future__ import annotations

import os
import socket
import struct
from pathlib import Path
from typing import BinaryIO

DEFAULT_PORT = 50000
MAX_FILE_BYTES = 100 * 1024 * 1024
MAX_FILE_NAME_BYTES = 1024
_BUFFER_SIZE = 64 * 1024


def send(
    file_path: str | os.PathLike[str],
    host: str,
    port: int = DEFAULT_PORT,
    timeout: float = 10.0,
) -> None:
    """Send a JSON or YAML file to the Go Wi-Fi receiver."""
    _validate_endpoint(host, port, timeout)
    source = Path(file_path)
    if not source.is_file():
        raise ValueError(f"source is not a regular file: {source}")
    name = source.name
    _validate_file_name(name)
    name_bytes = name.encode("utf-8")
    if len(name_bytes) > MAX_FILE_NAME_BYTES:
        raise ValueError(f"file name exceeds {MAX_FILE_NAME_BYTES} UTF-8 bytes")

    size = source.stat().st_size
    if size > MAX_FILE_BYTES:
        raise ValueError(f"file exceeds the {MAX_FILE_BYTES}-byte transfer limit")

    with socket.create_connection((host, port), timeout=timeout) as connection:
        connection.settimeout(timeout)
        connection.sendall(struct.pack(">I", len(name_bytes)))
        connection.sendall(name_bytes)
        connection.sendall(struct.pack(">Q", size))
        with source.open("rb") as file:
            _send_exact_size(connection, file, size)


def _send_exact_size(connection: socket.socket, file: BinaryIO, size: int) -> None:
    remaining = size
    while remaining:
        chunk = file.read(min(_BUFFER_SIZE, remaining))
        if not chunk:
            raise OSError("source file changed while it was being sent")
        connection.sendall(chunk)
        remaining -= len(chunk)


def _validate_file_name(name: str) -> None:
    if not name or name in {".", ".."} or any(char in name for char in ("/", "\\", "\0", ":")):
        raise ValueError("file name must not contain a path")
    if Path(name).suffix.lower() not in {".json", ".yaml", ".yml"}:
        raise ValueError("file name must end in .json, .yaml, or .yml")


def _validate_endpoint(host: str, port: int, timeout: float) -> None:
    if not host:
        raise ValueError("host must not be empty")
    if not 1 <= port <= 65535:
        raise ValueError("port must be between 1 and 65535")
    if timeout <= 0:
        raise ValueError("timeout must be positive")
