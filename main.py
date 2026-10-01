#!/usr/bin/env python3
"""Backwards-compatible launcher (``python main.py``). Prefer ``python -m vizcacha``."""

from vizcacha.ui.app import main

if __name__ == "__main__":
    raise SystemExit(main())
