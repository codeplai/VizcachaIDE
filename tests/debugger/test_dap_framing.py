import json

import pytest

from vizcacha.application.errors import DebugAdapterError
from vizcacha.infrastructure.delve_dap.framing import DapFrameDecoder, encode_message


def _wire(messages: list[dict]) -> bytes:
    return b"".join(encode_message(message) for message in messages)


def test_encode_uses_content_length_in_bytes():
    frame = encode_message({"seq": 1, "type": "request", "command": "ñandú"})

    header, body = frame.split(b"\r\n\r\n", 1)
    assert header == b"Content-Length: %d" % len(body)
    assert json.loads(body)["command"] == "ñandú"


def test_decoder_returns_every_message_of_a_recorded_session(functions_transcript):
    decoded = DapFrameDecoder().feed(_wire(functions_transcript))

    assert decoded == functions_transcript


def test_decoder_handles_messages_split_byte_by_byte(functions_transcript):
    decoder = DapFrameDecoder()
    wire = _wire(functions_transcript[:3])

    decoded = [
        message for index in range(len(wire)) for message in decoder.feed(wire[index : index + 1])
    ]

    assert decoded == functions_transcript[:3]


def test_decoder_waits_for_the_rest_of_the_body():
    decoder = DapFrameDecoder()
    wire = encode_message({"seq": 7, "type": "event", "event": "initialized"})

    assert decoder.feed(wire[:-5]) == []
    assert decoder.feed(wire[-5:]) == [{"seq": 7, "type": "event", "event": "initialized"}]


def test_header_name_is_case_insensitive_and_extra_headers_are_ignored():
    body = b'{"seq":1}'
    wire = b"content-length: %d\r\nContent-Type: application/json\r\n\r\n" % len(body) + body

    assert DapFrameDecoder().feed(wire) == [{"seq": 1}]


def test_missing_content_length_is_a_typed_error():
    with pytest.raises(DebugAdapterError):
        DapFrameDecoder().feed(b"Content-Type: x\r\n\r\n{}")


def test_invalid_json_is_a_typed_error():
    with pytest.raises(DebugAdapterError):
        DapFrameDecoder().feed(b"Content-Length: 3\r\n\r\n{x}")
