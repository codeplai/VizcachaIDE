"""Console of VizcachaIDE: one JSON line in, one JSON line out (docs/PLAN_PYTHON.md section 4.7).

Request:  {"code": "x = 5"}
Answer:   {"result": "", "output": "", "error": ""}

The session is one namespace that lives as long as the process. The Go side starts this script
with `python -X utf8 -u repl.py` and kills it on Reset or timeout.
"""
import builtins
import code
import contextlib
import io
import json
import sys
import traceback

# The Go side recognises this name in an error and shows its own translated text instead.
NO_INPUT_MARKER = "VizcachaConsoleNoInput"


class VizcachaConsoleNoInput(Exception):
    """input() was called in the console, which has no keyboard."""


class NoKeyboard:
    """Replaces sys.stdin for the code of the student: the real stdin carries the protocol."""

    def read(self, *args):
        raise VizcachaConsoleNoInput("the console cannot read the keyboard")

    readline = readlines = __next__ = read

    def __iter__(self):
        return self

    def isatty(self):
        return False


def no_input(*args):
    raise VizcachaConsoleNoInput("the console cannot read the keyboard")


def without_own_frames(tb):
    """Skips the leading frames that belong to this script."""
    while tb is not None and tb.tb_frame.f_code.co_filename == __file__:
        tb = tb.tb_next
    return tb


def format_error(etype, value, tb):
    return "".join(traceback.format_exception(etype, value, without_own_frames(tb))).rstrip()


def compile_snippet(source):
    """Returns (code object, is_expression). Raises SyntaxError when it is neither."""
    try:
        return compile(source, "<console>", "eval"), True
    except SyntaxError:
        pass  # not an expression; compiled below, outside the except so errors are not chained
    return compile(source, "<console>", "exec"), False


def run(source, namespace):
    """Runs a snippet; returns (result, output, error)."""
    buffer = io.StringIO()
    result = error = ""
    with contextlib.redirect_stdout(buffer), contextlib.redirect_stderr(buffer):
        try:
            compiled, is_expression = compile_snippet(source)
            value = eval(compiled, namespace) if is_expression else exec(compiled, namespace)
            if value is not None:
                result = repr(value)
        except BaseException:
            error = format_error(*sys.exc_info())
    return result, buffer.getvalue(), error


def serve(requests, answers):
    interpreter = code.InteractiveInterpreter({"__name__": "__main__", "__doc__": None})
    for line in requests:
        if not line.strip():
            continue
        try:
            source = json.loads(line)["code"]
            result, output, error = run(source, interpreter.locals)
        except (ValueError, KeyError, TypeError) as problem:
            result, output, error = "", "", "bad request: %s" % problem
        answers.write(json.dumps({"result": result, "output": output, "error": error}, ensure_ascii=False) + "\n")
        answers.flush()


def main():
    requests, answers = sys.stdin, sys.stdout
    sys.stdin = NoKeyboard()
    builtins.input = no_input
    serve(requests, answers)


if __name__ == "__main__":
    main()
