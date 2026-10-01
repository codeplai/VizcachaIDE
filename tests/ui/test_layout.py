"""Default panel layout, View > Reset Layout and window state persistence."""

from PyQt5.QtCore import Qt
from PyQt5.QtWidgets import QDockWidget, QLabel, QTabBar

from vizcacha.application.settings_keys import SettingsKeys


def _dock(workbench, panel_id: str) -> QDockWidget:
    return workbench.window.findChild(QDockWidget, panel_id)


def _area(workbench, panel_id: str):
    return workbench.window.dockWidgetArea(_dock(workbench, panel_id))


def _show(workbench, qtbot) -> None:
    workbench.window.resize(1280, 800)
    workbench.window.show()
    qtbot.waitExposed(workbench.window)
    _settle(workbench, qtbot)


def _settle(workbench, qtbot) -> None:
    """Dock geometries are updated by the layout on the next event loop iterations."""
    console, assistant = _dock(workbench, "console"), _dock(workbench, "assistant")
    qtbot.waitUntil(
        lambda: (
            console.geometry().right() < assistant.geometry().left()
            and console.geometry().top() == assistant.geometry().top()
        )
    )


def _current_tab(workbench, member: str) -> str | None:
    """Current tab text of the dock tab bar that contains a tab called ``member``."""
    for bar in workbench.window.findChildren(QTabBar):
        texts = [bar.tabText(index) for index in range(bar.count())]
        if member in texts and bar.isVisible():
            return bar.tabText(bar.currentIndex())
    return None


def _reset_action(workbench):
    return next(a for a in workbench.menu("view").actions() if a.objectName() == "reset_layout")


def test_default_layout_areas(workbench):
    assert _area(workbench, "project_files") == Qt.LeftDockWidgetArea
    assert _area(workbench, "variables") == Qt.RightDockWidgetArea
    assert _area(workbench, "callstack") == Qt.RightDockWidgetArea
    assert _area(workbench, "console") == Qt.BottomDockWidgetArea
    assert _area(workbench, "assistant") == Qt.BottomDockWidgetArea
    for panel_id in ("project_files", "variables", "callstack", "console", "assistant"):
        assert _dock(workbench, panel_id).isVisibleTo(workbench.window)


def test_right_panels_are_stacked_and_bottom_panels_side_by_side(workbench, qtbot):
    _show(workbench, qtbot)
    window = workbench.window

    variables, callstack = _dock(workbench, "variables"), _dock(workbench, "callstack")
    console, assistant = _dock(workbench, "console"), _dock(workbench, "assistant")

    assert window.tabifiedDockWidgets(variables) == []
    assert variables.geometry().bottom() < callstack.geometry().top()
    assert window.tabifiedDockWidgets(console) == []
    assert console.geometry().right() < assistant.geometry().left()
    assert assistant.width() >= window.width() // 3  # the Assistant is not squeezed
    qtbot.waitUntil(lambda: assistant.height() >= window.height() * 0.28)  # sizes applied


def test_left_panels_share_tabs_and_files_stays_in_front(workbench, qtbot):
    _show(workbench, qtbot)
    outline = workbench.add_panel("outline_like", "Outline", QLabel("symbols"), "left")
    files = _dock(workbench, "project_files")

    assert workbench.window.tabifiedDockWidgets(files) == [outline]
    qtbot.waitUntil(lambda: _current_tab(workbench, "Outline") == "Files")


def test_panels_of_the_same_group_become_tabs(workbench):
    first = workbench.add_panel("first", "First", QLabel("1"), "right", group="extra")
    second = workbench.add_panel("second", "Second", QLabel("2"), "right", group="extra")

    assert workbench.window.tabifiedDockWidgets(first) == [second]


def test_reset_layout_restores_places_and_visibility(workbench, qtbot):
    _show(workbench, qtbot)
    window = workbench.window
    assistant, variables = _dock(workbench, "assistant"), _dock(workbench, "variables")
    window.addDockWidget(Qt.LeftDockWidgetArea, assistant)
    variables.hide()

    _reset_action(workbench).trigger()

    assert window.dockWidgetArea(assistant) == Qt.BottomDockWidgetArea
    assert variables.isVisible()
    _settle(workbench, qtbot)


def test_window_state_is_saved_on_close_and_restored(qtbot, settings):
    from vizcacha.ui.app import build_services, build_workbench

    first = build_workbench(build_services(settings))
    qtbot.addWidget(first.window)
    _show(first, qtbot)
    first.window.resize(1100, 700)
    _dock(first, "variables").hide()
    first.window.close()
    assert settings.get(SettingsKeys.WINDOW_STATE, "")
    assert settings.get(SettingsKeys.WINDOW_GEOMETRY, "")

    second = build_workbench(build_services(settings))
    qtbot.addWidget(second.window)

    assert not _dock(second, "variables").isVisibleTo(second.window)
    assert _dock(second, "assistant").isVisibleTo(second.window)
    assert second.window.size().width() == 1100


def test_corrupt_saved_state_falls_back_to_default_layout(qtbot, settings):
    from vizcacha.ui.app import build_services, build_workbench

    settings.set(SettingsKeys.WINDOW_STATE, "not a layout")
    settings.set(SettingsKeys.WINDOW_GEOMETRY, 42)
    workbench = build_workbench(build_services(settings))
    qtbot.addWidget(workbench.window)

    assert _area(workbench, "assistant") == Qt.BottomDockWidgetArea


def test_status_bar_is_visible_by_default_with_indicators(workbench):
    bar = workbench.window.statusBar()

    assert not bar.isHidden()
    assert bar.findChild(QLabel, "status_language").text() == "English"
    assert bar.findChild(QLabel, "status_go_version").text()
