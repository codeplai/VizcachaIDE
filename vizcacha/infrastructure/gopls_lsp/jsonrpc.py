"""JSON-RPC 2.0 framing used by the Language Server Protocol (``Content-Length`` headers).

Pure Python, no Qt and no process: it is tested with recorded transcripts.
"""

import json

from vizcacha.application.errors import LanguageServerError

HEADER_SEPARATOR = b"\r\n\r\n"
LENGTH_HEADER = "content-length"
ENCODING = "utf-8"


class LanguageServerTimeoutError(LanguageServerError):
    """The server did not answer in time (gopls may still be loading packages)."""


def encode_message(message: dict) -> bytes:
    """Serialize one JSON-RPC message with its ``Content-Length`` header."""
    body = json.dumps(message, separators=(",", ":"), ensure_ascii=False).encode(ENCODING)
    return f"Content-Length: {len(body)}\r\n\r\n".encode("ascii") + body


def _content_length(header_block: bytes) -> int:
    for line in header_block.decode("ascii", errors="replace").split("\r\n"):
        name, separator, value = line.partition(":")
        if separator and name.strip().lower() == LENGTH_HEADER:
            try:
                return int(value.strip())
            except ValueError as error:
                raise LanguageServerError(f"Invalid Content-Length: {value!r}") from error
    raise LanguageServerError("Missing Content-Length header")


class MessageReader:
    """Accumulates bytes from the server and yields complete JSON messages."""

    def __init__(self) -> None:
        self._buffer = b""

    def feed(self, data: bytes) -> list[dict]:
        """Add ``data`` and return every message completed by it (possibly none).

        Raises LanguageServerError if the stream is corrupt.
        """
        self._buffer += data
        messages = []
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
        if len(self._buffer) < body_start + length:
            return None
        body = self._buffer[body_start : body_start + length]
        self._buffer = self._buffer[body_start + length :]
        try:
            return json.loads(body.decode(ENCODING))
        except (UnicodeDecodeError, json.JSONDecodeError) as error:
            raise LanguageServerError(f"Invalid JSON-RPC body: {error}") from error
