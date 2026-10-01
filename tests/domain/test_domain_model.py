from pathlib import Path

from vizcacha.domain.completion import CompletionItem, CompletionKind
from vizcacha.domain.debugging import DebugState, StackFrame, StopReason, Variable
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration, RunTarget


def test_run_configuration_for_file_runs_the_file_from_its_folder():
    config = RunConfiguration.for_file(Path("/work/hello.go"), ("a", "b"))

    assert config.working_dir == Path("/work")
    assert config.go_target_argument() == "hello.go"
    assert config.program_args == ("a", "b")
    assert config.executable_name(windows=True) == "hello.exe"
    assert config.executable_name(windows=False) == "hello"


def test_package_run_targets_current_directory():
    config = RunConfiguration(Path("/work/app"), Path("/work/app"), mode=RunTarget.PACKAGE)

    assert config.go_target_argument() == "."
    assert config.executable_name(windows=False) == "app"


def test_variable_children_are_eager_or_lazy():
    assert not Variable("x", "int", "1").has_children
    assert Variable("s", "[]int", "len: 2", reference=7).has_children
    assert Variable("p", "Point", "{}", children=(Variable("X", "int", "0"),)).has_children


def test_debug_state_current_location_is_top_frame():
    top = SourceLocation(Path("main.go"), 12)
    frames = (StackFrame(1, "main.add", top), StackFrame(2, "main.main", None))

    assert DebugState(StopReason.STEP, frames).current_location == top
    assert DebugState(StopReason.ENTRY, ()).current_location is None


def test_completion_inserts_label_unless_insert_text_given():
    assert CompletionItem("Println", CompletionKind.FUNCTION).text_to_insert == "Println"
    item = CompletionItem("for", CompletionKind.KEYWORD, insert_text="for {}")
    assert item.text_to_insert == "for {}"
