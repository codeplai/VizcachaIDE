"""Default arrangement of the dock panels, inspired by Thonny.

Each area has a rule, so features only say *where* their panel goes:

* ``left``   -> panels are tabs of one group (Files | Outline).
* ``right``  -> panels are stacked from top to bottom (Variables over Call Stack).
* ``bottom`` -> panels sit side by side (Console | Assistant).

A panel with an explicit ``group`` becomes a tab of the first panel of that group.
Sizes are fractions of the window, so the layout works from 1024x640 upwards.
"""

from dataclasses import dataclass

from PyQt5.QtCore import Qt
from PyQt5.QtWidgets import QDockWidget, QMainWindow

PANEL_AREAS = {
    "right": Qt.RightDockWidgetArea,
    "bottom": Qt.BottomDockWidgetArea,
    "left": Qt.LeftDockWidgetArea,
}
TABBED_AREAS = frozenset({"left"})
SPLIT_ORIENTATION = {"right": Qt.Vertical, "bottom": Qt.Horizontal}
# (fraction of the window, minimum pixels): width for left/right, height for bottom.
AREA_SIZES = {"left": (0.18, 190), "right": (0.22, 220), "bottom": (0.30, 180)}


@dataclass
class PanelPlacement:
    dock: QDockWidget
    area: str
    group: str | None
    visible: bool = True


class PanelLayout:
    def __init__(self, window: QMainWindow) -> None:
        self._window = window
        self._placements: list[PanelPlacement] = []
        window.setCorner(Qt.BottomLeftCorner, Qt.BottomDockWidgetArea)
        window.setCorner(Qt.BottomRightCorner, Qt.BottomDockWidgetArea)

    def place(
        self, dock: QDockWidget, area: str, group: str | None = None, dock_now: bool = True
    ) -> None:
        """``dock_now=False`` only remembers the default place (the dock is already docked)."""
        group = group or (area if area in TABBED_AREAS else None)
        placement = PanelPlacement(dock, area, group)
        if dock_now:
            self._dock(placement)
        self._placements.append(placement)

    def placement(self, dock: QDockWidget) -> PanelPlacement | None:
        return next((item for item in self._placements if item.dock is dock), None)

    def remember_visibility(self) -> None:
        """The default visibility is the one each panel has when the IDE starts."""
        for placement in self._placements:
            placement.visible = placement.dock.isVisibleTo(self._window)

    def reset(self) -> None:
        for placement in self._placements:
            self._window.removeDockWidget(placement.dock)
            placement.dock.setFloating(False)
        placed, self._placements = self._placements, []
        for placement in placed:
            self._dock(placement)
            self._placements.append(placement)
        for placement in placed:
            placement.dock.setVisible(placement.visible)
        self.raise_group_leaders()
        self.apply_sizes()

    def raise_group_leaders(self) -> None:
        leaders: dict[str, QDockWidget] = {}
        for placement in self._placements:
            if placement.group and placement.visible:
                leaders.setdefault(placement.group, placement.dock)
        for dock in leaders.values():
            dock.raise_()

    def apply_sizes(self) -> None:
        width, height = self._window.width(), self._window.height()
        for area, (fraction, minimum) in AREA_SIZES.items():
            docks = self._visible_docks(area)
            if not docks:
                continue
            across = Qt.Vertical if area == "bottom" else Qt.Horizontal
            total = height if area == "bottom" else width
            size = max(minimum, int(total * fraction))
            self._window.resizeDocks(docks, [size] * len(docks), across)
            self._share_evenly(area, docks)

    def _share_evenly(self, area: str, docks: list[QDockWidget]) -> None:
        if area not in SPLIT_ORIENTATION or len(docks) < 2:
            return
        along = SPLIT_ORIENTATION[area]
        total = self._window.width() if along == Qt.Horizontal else self._window.height()
        self._window.resizeDocks(docks, [total // len(docks)] * len(docks), along)

    def _visible_docks(self, area: str) -> list[QDockWidget]:
        return [
            item.dock
            for item in self._placements
            if item.area == area and item.dock.isVisibleTo(self._window)
        ]

    def _dock(self, placement: PanelPlacement) -> None:
        dock, area = placement.dock, placement.area
        dock.show()
        leader = self._group_leader(placement.group)
        if leader is not None:
            self._window.tabifyDockWidget(leader, dock)
            leader.raise_()  # the first panel of a group stays in front
            return
        previous = self._last_in_area(area)
        self._window.addDockWidget(PANEL_AREAS[area], dock)
        if previous is not None and area in SPLIT_ORIENTATION:
            self._window.splitDockWidget(previous, dock, SPLIT_ORIENTATION[area])

    def _group_leader(self, group: str | None) -> QDockWidget | None:
        if group is None:
            return None
        return next((item.dock for item in self._placements if item.group == group), None)

    def _last_in_area(self, area: str) -> QDockWidget | None:
        docks = [item.dock for item in self._placements if item.area == area]
        return docks[-1] if docks else None
