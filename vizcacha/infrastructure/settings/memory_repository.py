"""In-memory SettingsRepository, used by tests and by headless tools."""

from typing import TypeVar

T = TypeVar("T")


class InMemorySettingsRepository:
    def __init__(self, values: dict[str, object] | None = None) -> None:
        self.values: dict[str, object] = dict(values or {})

    def get(self, key: str, default: T) -> T:
        return self.values.get(key, default)  # type: ignore[return-value]

    def set(self, key: str, value: object) -> None:
        self.values[key] = value
