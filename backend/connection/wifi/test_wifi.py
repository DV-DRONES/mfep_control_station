import socket
import struct
import sys
import threading
import unittest
from pathlib import Path
from tempfile import TemporaryDirectory

project_root = Path(__file__).resolve().parents[3]
if str(project_root) not in sys.path:
    sys.path.insert(0, str(project_root))

from backend.connection import wifi


class WifiTransferTests(unittest.TestCase):
    def test_send_writes_compatible_packet(self) -> None:
        with TemporaryDirectory() as temporary_directory:
            source = Path(temporary_directory) / "settings.yaml"
            content = b"device:\n  mode: test\n"
            source.write_bytes(content)

            with socket.socket() as listener:
                listener.bind(("127.0.0.1", 0))
                listener.listen(1)
                port = listener.getsockname()[1]
                received: list[bytes] = []

                def accept_packet() -> None:
                    connection, _ = listener.accept()
                    with connection:
                        chunks = []
                        while chunk := connection.recv(4096):
                            chunks.append(chunk)
                        received.append(b"".join(chunks))

                thread = threading.Thread(target=accept_packet)
                thread.start()
                wifi.send(source, host="127.0.0.1", port=port)
                thread.join(timeout=5)

            self.assertFalse(thread.is_alive())
            packet = received[0]
            name_length = struct.unpack(">i", packet[:4])[0]
            name_end = 4 + name_length
            name = packet[4:name_end].decode("utf-8")
            content_size = struct.unpack(">q", packet[name_end : name_end + 8])[0]
            self.assertEqual(name, "settings.yaml")
            self.assertEqual(content_size, len(content))
            self.assertEqual(packet[name_end + 8 :], content)

    def test_rejects_traversal_and_unsupported_extensions(self) -> None:
        for name in ("../outside.json", r"..\outside.yaml", "notes.txt"):
            with self.subTest(name=name), self.assertRaises(ValueError):
                wifi._validate_file_name(name)


if __name__ == "__main__":
    unittest.main()
