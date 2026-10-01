"""JSON-RPC framing (Content-Length) with a transcript recorded from gopls v0.23."""

import json
from pathlib import Path

import pytest

from vizcacha.application.errors import LanguageServerError
from vizcacha.infrastructure.gopls_lsp.jsonrpc import MessageReader, encode_message

TRANSCRIPT = Path(__file__).parent / "transcripts" / "gopls_session.json"


def recorded_messages() -> list[dict]:
    return list(json.loads(TRANSCRIPT.read_text(encoding="utf-8")).values())


def test_encode_message_writes_content_length_in_bytes():
    frame = encode_message({"jsonrpc": "2.0", "method": "x", "params": {"text": "ñandú"}})

    header, body = frame.split(b"\r\n\r\n", 1)
    assert header == f"Content-Length: {len(body)}".encode("ascii")
    assert len(body) > len(body.decode("utf-8"))  # multi-byte characters counted as bytes


def test_reader_round_trips_a_recorded_session_split_in_small_chunks():
    stream = b"".join(encode_message(message) for message in recorded_messages())
    reader = MessageReader()

    received = []
    for start in range(0, len(stream), 7):
        received += reader.feed(stream[start : start + 7])

    assert received == recorded_messages()


def test_reader_returns_several_messages_from_one_chunk():
    first, second = {"id": 1, "result": None}, {"method": "initialized", "params": {}}
    reader = MessageReader()

    assert reader.feed(encode_message(first) + encode_message(second)) == [first, second]
    assert reader.feed(b"") == []


def test_reader_waits_for_the_rest_of_the_body():
    frame = encode_message({"id": 7, "result": [1, 2, 3]})
    reader = MessageReader()

    assert reader.feed(frame[:-3]) == []
    assert reader.feed(frame[-3:]) == [{"id": 7, "result": [1, 2, 3]}]


def test_reader_accepts_extra_headers_in_any_case():
    body = b'{"id":1,"result":true}'
    frame = b"content-type: application/vscode-jsonrpc\r\ncontent-length: %d\r\n\r\n" % len(body)

    assert MessageReader().feed(frame + body) == [{"id": 1, "result": True}]


@pytest.mark.parametrize(
    "frame",
    [
        b"Content-Type: x\r\n\r\n{}",
        b"Content-Length: abc\r\n\r\n{}",
        b"Content-Length: 5\r\n\r\n{nop}",
    ],
)
def test_corrupt_streams_raise_language_server_error(frame):
    with pytest.raises(LanguageServerError):
        MessageReader().feed(frame)
