"""Translate DAP response bodies (dicts) into domain debugging objects.

Pure functions, tested with transcripts recorded from ``dlv dap`` 1.27.
"""

import re
from pathlib import Path

from vizcacha.domain.debugging import Goroutine, StackFrame, StopReason, Variable
from vizcacha.domain.diagnostics import SourceLocation

MAX_VALUE_LENGTH = 200
ELLIPSIS = "…"
LOCALS_SCOPE_PREFIX = "Locals"
SUBTLE_HINT = "subtle"  # Delve marks runtime frames (e.g. inside panic) as "subtle"

STOP_REASONS = {
    "breakpoint": StopReason.BREAKPOINT,
    "function breakpoint": StopReason.BREAKPOINT,
    "data breakpoint": StopReason.BREAKPOINT,
    "instruction breakpoint": StopReason.BREAKPOINT,
    "step": StopReason.STEP,
    "goto": StopReason.STEP,
    "entry": StopReason.ENTRY,
    "exception": StopReason.PANIC,
    "panic": StopReason.PANIC,
    "pause": StopReason.PAUSE,
}
OUTPUT_CATEGORIES = {"stdout": "stdout", "stderr": "stderr"}
NOISE_OUTPUT = ("Type 'dlv help' for list of commands.",)
CURRENT_THREAD_MARK = re.compile(r"^\*\s*")


def stop_reason(reason: str) -> StopReason:
    return STOP_REASONS.get(reason, StopReason.PAUSE)


def stop_description(body: dict) -> str:
    parts = [body.get("description", ""), body.get("text", "")]
    return ": ".join(part.strip('"') for part in parts if part)


def truncate_value(value: str, limit: int = MAX_VALUE_LENGTH) -> str:
    if len(value) <= limit:
        return value
    return value[: limit - 1] + ELLIPSIS


def map_variable(raw: dict) -> Variable:
    return Variable(
        name=raw.get("name", ""),
        type_name=raw.get("type", ""),
        value=truncate_value(raw.get("value", "")),
        reference=int(raw.get("variablesReference", 0) or 0),
    )


def map_variables(body: dict) -> list[Variable]:
    return [map_variable(raw) for raw in body.get("variables", [])]


def _location(raw: dict) -> SourceLocation | None:
    path = (raw.get("source") or {}).get("path")
    line = raw.get("line", 0)
    if not path or line <= 0:
        return None
    return SourceLocation(Path(path), line, max(raw.get("column", 1), 1))


def map_frame(raw: dict) -> StackFrame:
    return StackFrame(frame_id=raw["id"], function=raw.get("name", ""), location=_location(raw))


def user_frames(raw_frames: list[dict]) -> list[dict]:
    """Drop the leading runtime frames (panic machinery) so the user sees their own code."""
    for index, raw in enumerate(raw_frames):
        if raw.get("presentationHint") != SUBTLE_HINT:
            return raw_frames[index:]
    return raw_frames


def map_frames(body: dict, skip_runtime: bool = False) -> tuple[StackFrame, ...]:
    raw_frames = body.get("stackFrames", [])
    if skip_runtime:
        raw_frames = user_frames(raw_frames)
    return tuple(map_frame(raw) for raw in raw_frames)


def locals_reference(body: dict) -> int:
    """``variablesReference`` of the Locals scope, 0 if there is none."""
    for scope in body.get("scopes", []):
        if scope.get("name", "").startswith(LOCALS_SCOPE_PREFIX):
            return int(scope.get("variablesReference", 0) or 0)
    return 0


def map_goroutines(body: dict) -> tuple[Goroutine, ...]:
    return tuple(
        Goroutine(goroutine_id=raw["id"], name=CURRENT_THREAD_MARK.sub("", raw.get("name", "")))
        for raw in body.get("threads", [])
    )


def output_category(category: str) -> str:
    """DAP categories → port categories: "stdout" | "stderr" | "console"."""
    return OUTPUT_CATEGORIES.get(category, "console")


def is_noise(output: str) -> bool:
    return output.strip() in NOISE_OUTPUT


def output_event(body: dict) -> tuple[str, str] | None:
    """(text, category) for an ``output`` event, or None when it is not for the user."""
    text, category = body.get("output", ""), body.get("category", "")
    if category == "telemetry" or not text or is_noise(text):
        return None
    return text, output_category(category)
