"""
java python converter
"""

from __future__ import annotations

import os
import shutil
import subprocess
from pathlib import Path

DEFAULT_PORT = 50000
_PROJECT_DIR = Path(__file__).resolve().parent


def _java_command() -> list[str]:
    java = os.environ.get("JAVA") or shutil.which("java")
    if not java:
        raise FileNotFoundError(
            "Java was not found."
        )

    jar_candidates = (
        _PROJECT_DIR / "build" / "wifi.jar",
        _PROJECT_DIR / "lib" / "wifi.jar",
    )
    jar_path = next((path for path in jar_candidates if path.is_file()), None)
    if jar_path is None:
        raise FileNotFoundError(
            "Wi-Fi JAR was not found."
            f"'.\\build.ps1' in '{_PROJECT_DIR}'."
        )
    return [java, "-jar", str(jar_path)]


def send(
    ip: str,
    file_path: str | os.PathLike[str],
    port: int = DEFAULT_PORT,
    timeout: float | None = 20.0,
) -> None:
    """.json / .yaml/ """
    command = _java_command() + ["send", ip, str(file_path), str(port)]
    subprocess.run(command, check=True, capture_output=True, text=True, timeout=timeout)


def receive(
    port: int = DEFAULT_PORT,
    save_dir: str | os.PathLike[str] = "received",
    timeout: float | None = None,
) -> Path:
    
    command = _java_command() + ["receive", str(port), str(save_dir)]
    result = subprocess.run(
        command, check=True, capture_output=True, text=True, timeout=timeout
    )
    prefix = "Received: "
    for line in result.stdout.splitlines():
        if line.startswith(prefix):
            return Path(line[len(prefix):])
    raise RuntimeError("Java Wi-Fi receiver exited without reporting the received file")


recive = receive

__all__ = ["DEFAULT_PORT", "send", "receive", "recive"]
