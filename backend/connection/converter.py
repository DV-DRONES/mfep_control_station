import subprocess
from pathlib import Path
from typing import Union

__all__ = ["Wifi"]


class Wifi:


  DEFAULT_PORT = 50000

  @classmethod
  def _get_bin_dir(cls) -> Path:
 
    return Path(__file__).resolve().parent.parent / "bin"

  @classmethod
  def send(
      cls,
      ip: str,
      file_path: Union[str, Path],
      port: int = DEFAULT_PORT,
  ) -> None:
    
    bin_dir = cls._get_bin_dir()
    resolved_file = Path(file_path).resolve()

    cmd = [
        "java",
        "-cp",
        str(bin_dir),
        "Wifi",
        "send",
        ip,
        str(resolved_file),
        str(port),
    ]

    try:
      result = subprocess.run(
          cmd, check=True, text=True, capture_output=True
      )
      if result.stdout:
        print(result.stdout.strip())
    except subprocess.CalledProcessError as e:
      error_msg = e.stderr.strip() if e.stderr else str(e)
      raise RuntimeError(f"sending error: {error_msg}") from e

  @classmethod
  def receive(
      cls, port: int = DEFAULT_PORT, save_directory: Union[str, Path] = "received"
  ) -> Path:
    
    bin_dir = cls._get_bin_dir()
    resolved_dir = Path(save_directory).resolve()

    cmd = [
        "java",
        "-cp",
        str(bin_dir),
        "Wifi",
        "receive",
        str(port),
        str(resolved_dir),
    ]

    try:
      result = subprocess.run(
          cmd, check=True, text=True, capture_output=True
      )
      output = result.stdout.strip()
      if output:
        print(output)


      if "Received: " in output:
        path_str = output.split("Received: ")[1].strip()
        return Path(path_str)

      return resolved_dir
    except subprocess.CalledProcessError as e:
      error_msg = e.stderr.strip() if e.stderr else str(e)
      raise RuntimeError(f"reciving error: {error_msg}") from e