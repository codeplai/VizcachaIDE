"""Debug Adapter Protocol wire format: ``Content-Length`` header + JSON body.

Pure Python (no Qt, no sockets) so it can be tested with recorded transcripts.
"""

import json

from vizcacha.application.errors import DebugAdapterError

HEADER_SEPARATOR = b"\r\n\r\n"
LENGTH_HEADER = b"content-length"


def encode_message(message: dict) -> bytes:
    body = json.dumps(message, separators=(",", ":")).encode("utf-8")
    return b"Content-Length: %d\r\n\r\n" % len(body) + body


def _content_length(header: bytes) -> int:
    for line in header.split(b"\r\n"):
        name, _sep, value = line.partition(b":")
        if name.strip().lower() == LENGTH_HEADER:
            return int(value.strip())
    raise DebugAdapterError(f"DAP message without Content-Length: {header!r}")


class DapFrameDecoder:
    """Accumulates bytes and returns every complete DAP message."""

    def __init__(self) -> None:
        self._buffer = b""

    def feed(self, data: bytes) -> list[dict]:
        self._buffer += data
        messages: list[dict] = []
        while True:
            message = self._next_message()
            if message is None:
                return messages
            messages.append(message)

    def _next_message(self) -> dict | None:
        header_end = self._buffer.find(HEADER_SEPARATOR)
        if header_end < 0:
            return None
        length = _content_length(self._buffer[:header_end])
        body_start = header_end + len(HEADER_SEPARATOR)
        body_end = body_start + length
        if len(self._buffer) < body_end:
            return None
        body = self._buffer[body_start:body_end]
        self._buffer = self._buffer[body_end:]
        try:
            return json.loads(body.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError) as error:
            raise DebugAdapterError(f"Invalid DAP message: {error}") from error
