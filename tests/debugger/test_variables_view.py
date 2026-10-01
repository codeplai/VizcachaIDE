"""VariablesView: asynchronous lazy expansion and expansion state kept across steps."""

import pytest

from vizcacha.domain.debugging import Variable
from vizcacha.ui.features.debugger import variables_view
from vizcacha.ui.features.debugger.variables_view import VariablesView, name_path


@pytest.fixture
def requested() -> list[int]:
    return []


@pytest.fixture
def view(qtbot, requested):
    widget = VariablesView(requested.append)
    qtbot.addWidget(widget)
    return widget


def _point(reference: int) -> Variable:
    return Variable("p", "main.Point", "{X: 1, Y: 2}", reference=reference)


def _children(item) -> list[str]:
    return [item.child(i).text(0) for i in range(item.childCount())]


def test_expanding_shows_loading_until_the_children_arrive(view, requested):
    view.show_variables([_point(7), Variable("n", "int", "1")])
    item = view.topLevelItem(0)
    assert requested == [] and item.childCount() == 0

    item.setExpanded(True)
    loading = _children(item)
    view.show_children(7, [Variable("X", "int", "1"), Variable("Y", "int", "2")])
    item.setExpanded(False)
    item.setExpanded(True)

    assert loading == ["Loading..."]
    assert requested == [7]
    assert _children(item) == ["X", "Y"]
    assert view.topLevelItem(1).childIndicatorPolicy() != item.ShowIndicator


def test_loading_text_goes_through_the_translator(view, monkeypatch):
    monkeypatch.setattr(variables_view, "_", str.upper)
    view.show_variables([_point(7)])

    view.topLevelItem(0).setExpanded(True)

    assert view.topLevelItem(0).child(0).text(0) == "LOADING..."
    assert view.topLevelItem(0).child(0).isDisabled()


def test_an_empty_answer_removes_the_arrow(view):
    view.show_variables([_point(7)])
    item = view.topLevelItem(0)
    item.setExpanded(True)

    view.show_children(7, [])

    assert item.childCount() == 0
    assert item.childIndicatorPolicy() == item.DontShowIndicatorWhenChildless


def test_unknown_or_outdated_answers_are_ignored(view):
    view.show_variables([_point(7)])
    view.topLevelItem(0).setExpanded(True)
    view.show_variables([_point(8)])  # the program stepped: the old item is gone

    view.show_children(7, [Variable("X", "int", "1")])
    view.show_children(99, [Variable("Z", "int", "1")])

    assert _children(view.topLevelItem(0)) == ["Loading..."]  # still waiting for 8


def test_expansion_is_restored_after_a_step(view, requested):
    view.show_variables([_point(7)])
    view.topLevelItem(0).setExpanded(True)
    view.show_children(7, [Variable("inner", "main.Inner", "{...}", reference=9)])
    view.topLevelItem(0).child(0).setExpanded(True)
    view.show_children(9, [Variable("deep", "int", "3")])

    view.show_variables([_point(17)])  # Delve renumbers the references at every stop
    view.show_children(17, [Variable("inner", "main.Inner", "{...}", reference=19)])
    view.show_children(19, [Variable("deep", "int", "4")])

    item = view.topLevelItem(0)
    assert requested == [7, 9, 17, 19]
    assert item.isExpanded() and item.child(0).isExpanded()
    assert item.child(0).child(0).text(2) == "4"
    assert name_path(item.child(0).child(0)) == ("p", "inner", "deep")


def test_collapsed_variables_stay_collapsed(view, requested):
    view.show_variables([_point(7)])
    view.topLevelItem(0).setExpanded(True)
    view.topLevelItem(0).setExpanded(False)

    view.show_variables([_point(17)])

    assert requested == [7]
    assert not view.topLevelItem(0).isExpanded()


def test_a_new_session_forgets_the_expansion(view, requested):
    view.show_variables([_point(7)])
    view.topLevelItem(0).setExpanded(True)

    view.forget_expansion()
    view.show_variables([_point(17)])

    assert requested == [7]
