"""Beginner-friendly explanations of Go errors, shown in English or Spanish."""

from dataclasses import dataclass, field


@dataclass(frozen=True)
class ErrorExplanation:
    """Explanation for a recognised error.

    ``explanation_id`` is stable across versions and languages (e.g. ``E-UNUSED-VAR``)
    so it can be used in tests, help links and bug reports.

    ``title``, ``body`` and ``fix_hint`` are English gettext message ids. The UI
    translates them at display time and then fills ``{placeholders}`` from
    ``placeholders`` (e.g. ``{"name": "x"}``).
    """

    explanation_id: str
    title: str
    body: str
    fix_hint: str = ""
    placeholders: dict[str, str] = field(default_factory=dict)
