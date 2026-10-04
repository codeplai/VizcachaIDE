"""Record real compiler and runtime output for the C++ error parser fixtures (track C0).

Usage: python record_cpp_output.py <llvm bin> <gcc bin> <fixtures folder>
Writes <fixtures>/{clang,gcc}/<case>.txt. Programs live in a folder with spaces and an accent,
like students' folders; compile flags are the adapter's (cpp.CompileFlags on Windows).
"""
import os
import subprocess
import sys
import tempfile

FLAGS = ["-std=c++17", "-g", "-O0", "-Wall", "-Wextra", "-fdiagnostics-color=never", "-static"]

COMPILE = {
    "undeclared": 'int main() {\n    int total = 1;\n    return totl;\n}\n',
    "cout_no_include": 'int main() {\n    std::cout << "hola";\n    return 0;\n}\n',
    "cout_no_std": '#include <iostream>\nint main() {\n    cout << "hola" << endl;\n    return 0;\n}\n',
    "missing_semicolon": '#include <iostream>\nint main() {\n    int x = 5\n    std::cout << x;\n    return 0;\n}\n',
    "missing_brace": '#include <iostream>\nint main() {\n    std::cout << "hola";\n    return 0;\n',
    "expected_primary": 'int main() {\n    int x = ;\n    return x;\n}\n',
    "no_matching_function": '#include <string>\nvoid saludar(std::string nombre, int veces) {}\nint main() {\n    saludar(5, "tres");\n    return 0;\n}\n',
    "too_few_args": 'int sumar(int a, int b) { return a + b; }\nint main() {\n    return sumar(1);\n}\n',
    "too_many_args": 'int doble(int a) { return 2 * a; }\nint main() {\n    return doble(1, 2);\n}\n',
    "cannot_convert": '#include <string>\nint main() {\n    std::string nombre = "Ana";\n    int n = nombre;\n    return n;\n}\n',
    "invalid_operands": '#include <string>\n#include <vector>\nint main() {\n    std::vector<int> v;\n    std::string s = "a";\n    auto r = v + s;\n    return 0;\n}\n',
    "no_member": '#include <string>\nint main() {\n    std::string s = "hola";\n    return s.lenght();\n}\n',
    "redeclared": 'int main() {\n    int x = 1;\n    int x = 2;\n    return x;\n}\n',
    "missing_return": 'int doble(int a) {\n    int r = 2 * a;\n}\nint main() {\n    return doble(2);\n}\n',
    "no_such_file": '#include <iostreem>\nint main() {\n    return 0;\n}\n',
    "unused_variable": 'int main() {\n    int sin_usar = 3;\n    return 0;\n}\n',
    "uninitialized": '#include <iostream>\nint main() {\n    int x;\n    std::cout << x;\n    return 0;\n}\n',
    "sign_compare": '#include <vector>\nint main() {\n    std::vector<int> v(3);\n    for (int i = 0; i < v.size(); i++) {}\n    return 0;\n}\n',
    "assign_in_condition": 'int main() {\n    int x = 1;\n    if (x = 2) {\n        return 1;\n    }\n    return 0;\n}\n',
    "array_bounds": 'int main() {\n    int a[3] = {1, 2, 3};\n    return a[5];\n}\n',
    "string_compare": 'int main() {\n    const char* s = "si";\n    if (s == "si") {\n        return 1;\n    }\n    return 0;\n}\n',
    "undefined_reference": 'int calcular(int x);\nint main() {\n    return calcular(2);\n}\n',
    "undefined_main": 'int ayudar() {\n    return 1;\n}\n',
}

RUNTIME = {
    "segfault": '#include <iostream>\nint main() {\n    std::cout << "antes de caer" << std::endl;\n    int* p = nullptr;\n    *p = 42;\n    return 0;\n}\n',
    "stack_overflow": 'int f(int n) {\n    volatile int a[1000];\n    a[0] = n;\n    return f(n + 1) + a[0];\n}\nint main() {\n    return f(0);\n}\n',
    "divide_zero": '#include <iostream>\nint main() {\n    int a = 10;\n    volatile int b = 0;\n    std::cout << a / b;\n    return 0;\n}\n',
    "terminate": '#include <stdexcept>\nint main() {\n    throw std::runtime_error("algo salio mal");\n}\n',
    "out_of_range": '#include <vector>\nint main() {\n    std::vector<int> v(3);\n    return v.at(10);\n}\n',
}


def header(folder, compiler, source, extra=""):
    lines = "".join(f"#   {line}\n" for line in source.splitlines())
    return f"# workingDir: {folder}\n# compiler: {compiler}\n{extra}# program:\n{lines}# output:\n"


def record(name, compiler, flags, out_dir):
    work = os.path.join(tempfile.mkdtemp(), "mis programas ñandú")
    os.makedirs(work)
    os.makedirs(out_dir, exist_ok=True)
    for case, source in {**COMPILE, **RUNTIME}.items():
        src = os.path.join(work, "main.cpp")
        exe = os.path.join(work, "main.exe")
        with open(src, "w", encoding="utf-8") as f:
            f.write(source)
        built = subprocess.run([compiler, *flags, "-o", exe, src], cwd=work, capture_output=True, text=True, encoding="utf-8", errors="replace")
        text = built.stdout + built.stderr
        extra = f"# compileExit: {built.returncode}\n"
        if case in RUNTIME and built.returncode == 0:
            ran = subprocess.run([exe], cwd=work, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=60)
            extra += f"# runExit: {ran.returncode} (0x{ran.returncode & 0xFFFFFFFF:08X})\n"
            text += ran.stdout + ran.stderr
        with open(os.path.join(out_dir, case + ".txt"), "w", encoding="utf-8", newline="\n") as f:
            f.write(header(work, name, source, extra) + text)
    print(f"{name}: {len(COMPILE) + len(RUNTIME)} cases in {out_dir}")


def main():
    llvm_bin, gcc_bin, out = sys.argv[1:4]
    exe = ".exe" if os.name == "nt" else ""
    record("clang++", os.path.join(llvm_bin, "clang++" + exe), FLAGS, os.path.join(out, "clang"))
    record("g++", os.path.join(gcc_bin, "g++" + exe), FLAGS, os.path.join(out, "gcc"))


main()
