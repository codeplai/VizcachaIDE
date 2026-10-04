"""Record real Python 3.12 output for the M1 error parser fixtures (track P0).

Usage: <python3.12> record_python_output.py <fixtures folder>
Each program runs in a folder whose name has spaces and an accent, like students' folders.
"""
import json
import os
import subprocess
import sys
import tempfile

PROGRAMS = {
    "name_error": 'print("hola")\nprint(nombre)\n',
    "nested_traceback": 'def promedio(notas):\n    return sum(notas) / len(notas)\n\ndef main():\n    print(promedio([]))\n\nmain()\n',
    "syntax_missing_colon": 'for i in range(3)\n    print(i)\n',
    "syntax_unclosed": 'print("hola"\nx = 1\n',
    "syntax_unterminated_string": 'print("hola)\n',
    "indentation_expected": 'def f():\nreturn 1\n',
    "indentation_unexpected": 'x = 1\n    y = 2\n',
    "tab_error": 'if True:\n\tx = 1\n        y = 2\n',
    "type_concat": 'edad = 20\nprint("Tienes " + edad + " años")\n',
    "value_literal": 'n = int("doce")\n',
    "index_range": 'lista = [1, 2, 3]\nprint(lista[5])\n',
    "key_error": 'd = {"a": 1}\nprint(d["b"])\n',
    "attribute_error": 'texto = "hola"\ntexto.push("!")\n',
    "module_not_found": 'import numpyy\n',
    "file_not_found": 'open("datos.txt")\n',
    "recursion": 'def f(n):\n    return f(n + 1)\n\nf(0)\n',
    "eof_input": 'nombre = input("Nombre: ")\n',
    "unbound_local": 'contador = 0\ndef sumar():\n    contador += 1\nsumar()\n',
    "missing_args": 'def saludar(nombre, edad):\n    pass\nsaludar("Ana")\n',
}

RUFF_SOURCE = 'import os\n\ndef main():\n    total = 5\n    print(valor)\n'


def run(python, folder, name, source, stdin=""):
    path = os.path.join(folder, name + ".py")
    with open(path, "w", encoding="utf-8") as f:
        f.write(source)
    env = dict(os.environ, PYTHONUTF8="1", PYTHONIOENCODING="utf-8", PYTHONUNBUFFERED="1",
               PYTHONDONTWRITEBYTECODE="1", NO_COLOR="1", PYTHON_COLORS="0")
    result = subprocess.run([python, "-X", "utf8", path], cwd=folder, input=stdin, capture_output=True,
                            text=True, encoding="utf-8", env=env, timeout=60)
    return result.stdout + result.stderr


def main():
    out_dir = sys.argv[1]
    os.makedirs(out_dir, exist_ok=True)
    python = sys.executable
    folder = os.path.join(tempfile.mkdtemp(), "mis programas ñandú")
    os.makedirs(folder)
    for name, source in PROGRAMS.items():
        text = run(python, folder, name, source)
        header = f"# workingDir: {folder}\n# program:\n" + "".join(f"#   {line}\n" for line in source.splitlines()) + "# output:\n"
        with open(os.path.join(out_dir, name + ".txt"), "w", encoding="utf-8", newline="\n") as f:
            f.write(header + text)
    missing = subprocess.run([python, "-X", "utf8", "no_existe.py"], cwd=folder, capture_output=True, text=True, encoding="utf-8")
    with open(os.path.join(out_dir, "cant_open_file.txt"), "w", encoding="utf-8", newline="\n") as f:
        f.write(f"# workingDir: {folder}\n# program: (none: python no_existe.py)\n# output:\n" + missing.stdout + missing.stderr)
    ruff_file = os.path.join(folder, "ruff_case.py")
    with open(ruff_file, "w", encoding="utf-8") as f:
        f.write(RUFF_SOURCE)
    ruff = subprocess.run([python, "-m", "ruff", "check", "--isolated", "--select", "E9,F", "--output-format", "json", "--no-cache", ruff_file],
                          cwd=folder, capture_output=True, text=True, encoding="utf-8")
    report = json.loads(ruff.stdout)
    with open(os.path.join(out_dir, "ruff_check.json"), "w", encoding="utf-8", newline="\n") as f:
        json.dump({"workingDir": folder, "source": RUFF_SOURCE, "report": report}, f, ensure_ascii=False, indent=2)
    print(f"recorded {len(PROGRAMS) + 2} fixtures with Python {sys.version.split()[0]} in {out_dir}")


main()
