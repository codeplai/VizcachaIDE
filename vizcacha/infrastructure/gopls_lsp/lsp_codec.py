"""LSP messages <-> JSON dictionaries, with lsprotocol types and its cattrs converter."""

from lsprotocol import types
from lsprotocol.converters import get_converter

from vizcacha.application.errors import LanguageServerError

CONVERTER = get_converter()


def request_payload(request_id: int, method: str, params: object) -> dict:
    if params is None:  # e.g. "shutdown"
        return {"jsonrpc": "2.0", "id": request_id, "method": method}
    request_cls = types.METHOD_TO_TYPES[method][0]
    return CONVERTER.unstructure(request_cls(id=request_id, params=params))


def notification_payload(method: str, params: object) -> dict:
    notification_cls = types.METHOD_TO_TYPES[method][0]
    kwargs = {} if params is None else {"params": params}
    return CONVERTER.unstructure(notification_cls(**kwargs))


def typed_result(message: dict, method: str) -> object:
    """Typed ``result`` of the response to ``method``. Raises LanguageServerError on errors."""
    if "error" in message:
        error = message["error"]
        raise LanguageServerError(str(error.get("message", error)))
    response_cls = types.METHOD_TO_TYPES[method][1]
    try:
        return CONVERTER.structure({"result": None, **message}, response_cls).result
    except (TypeError, ValueError, KeyError) as error:
        raise LanguageServerError(f"Unexpected {method} result: {error}") from error
