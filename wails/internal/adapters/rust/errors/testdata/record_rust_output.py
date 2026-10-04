"""Record real rustc, cargo and clippy output for the Rust error parser fixtures (track R0).

Usage: python record_rust_output.py <cargo home> <rustup home> <fixtures folder> [case ...]
Writes <fixtures>/{json,text}/<case>.txt (rustc on a loose file, with --error-format=json and
without), <fixtures>/runtime/<case>.txt (a program that panics, RUST_BACKTRACE=1),
<fixtures>/cargo/<case>.txt (cargo build/run --message-format=json in a project) and
<fixtures>/clippy/<case>.txt (clippy-driver on a loose file). Programs live in a folder with spaces
and an accent, like students' folders. Flags are the adapter's (docs/PLAN_RUST.md section 4.3).
"""
import os
import shutil
import subprocess
import sys
import tempfile

COMPILE = {
    "moved": 'fn main() {\n    let s = String::from("hola");\n    let t = s;\n    println!("{} {}", s, t);\n}\n',
    "borrow_mut_twice": "fn main() {\n    let mut v = vec![1];\n    let a = &mut v;\n    let b = &mut v;\n    a.push(2);\n    b.push(3);\n}\n",
    "borrow_conflict": "fn main() {\n    let mut v = vec![1, 2];\n    for x in &v {\n        v.push(*x);\n    }\n}\n",
    "assign_borrowed": "fn main() {\n    let mut x = 5;\n    let r = &x;\n    x = 6;\n    println!(\"{}\", r);\n}\n",
    "move_borrowed": "fn main() {\n    let s = String::from(\"a\");\n    let r = &s;\n    let t = s;\n    println!(\"{} {}\", r, t);\n}\n",
    "move_out_of_ref": "fn main() {\n    let v = vec![String::from(\"a\")];\n    let s = v[0];\n    println!(\"{}\", s);\n}\n",
    "dangling_ref": "fn main() {\n    let r;\n    {\n        let x = 5;\n        r = &x;\n    }\n    println!(\"{}\", r);\n}\n",
    "return_local_ref": "fn crear() -> &'static str {\n    let s = String::from(\"hola\");\n    &s\n}\nfn main() {\n    println!(\"{}\", crear());\n}\n",
    "missing_lifetime": "fn mayor(a: &str, b: &str) -> &str {\n    if a.len() > b.len() { a } else { b }\n}\nfn main() {\n    println!(\"{}\", mayor(\"a\", \"bb\"));\n}\n",
    "mismatched_types": "fn main() {\n    let n: i32 = \"cinco\";\n    println!(\"{}\", n);\n}\n",
    "missing_return": "fn doble(x: i32) -> i32 {\n    x * 2;\n}\nfn main() {\n    println!(\"{}\", doble(2));\n}\n",
    "unresolved_name": "fn main() {\n    let total = 1;\n    println!(\"{}\", totl);\n}\n",
    "failed_resolve": "fn main() {\n    let m: HashMap<i32, i32> = HashMap::new();\n    println!(\"{}\", m.len());\n}\n",
    "unresolved_import": "use rand::Rng;\nfn main() {\n    println!(\"hola\");\n}\n",
    "no_method": "fn main() {\n    let v = vec![1, 2];\n    println!(\"{}\", v.lenght());\n}\n",
    "no_field": "struct Punto { x: i32 }\nfn main() {\n    let p = Punto { x: 1 };\n    println!(\"{}\", p.y);\n}\n",
    "assign_twice": "fn main() {\n    let x = 5;\n    x = 6;\n    println!(\"{}\", x);\n}\n",
    "not_mutable": "fn main() {\n    let v = Vec::new();\n    v.push(1);\n    println!(\"{:?}\", v);\n}\n",
    "trait_bound": "struct Punto { x: i32 }\nfn main() {\n    let p = Punto { x: 1 };\n    println!(\"{}\", p);\n}\n",
    "question_mark": "fn main() {\n    let n: i32 = \"5\".parse()?;\n    println!(\"{}\", n);\n}\n",
    "wrong_args": "fn sumar(a: i32, b: i32) -> i32 { a + b }\nfn main() {\n    println!(\"{}\", sumar(1));\n}\n",
    "binary_op": "fn main() {\n    let a = \"ho\";\n    let b = \"la\";\n    println!(\"{}\", a + b);\n}\n",
    "deref": "fn main() {\n    let x = 5;\n    println!(\"{}\", *x);\n}\n",
    "non_exhaustive": "enum Color { Rojo, Verde, Azul }\nfn main() {\n    let c = Color::Rojo;\n    match c {\n        Color::Rojo => println!(\"r\"),\n        Color::Verde => println!(\"v\"),\n    }\n}\n",
    "private": "mod banco {\n    fn saldo() -> i32 { 5 }\n}\nfn main() {\n    println!(\"{}\", banco::saldo());\n}\n",
    "type_annotations": "fn main() {\n    let v = \"1 2\".split(' ').map(|s| s.parse().unwrap()).collect();\n    println!(\"{:?}\", v);\n}\n",
    "missing_items": "trait Forma { fn area(&self) -> f64; }\nstruct Cuadrado;\nimpl Forma for Cuadrado {}\nfn main() {}\n",
    "no_main": "fn ayudar() {}\n",
    "expected_token": "fn main() {\n    let x = 5\n    println!(\"{}\", x);\n}\n",
    "unclosed_delimiter": "fn main() {\n    println!(\"hola\");\n",
    "unknown_macro": "fn main() {\n    printn!(\"hola\");\n}\n",
    "warnings": "use std::collections::HashMap;\nfn nunca() {}\nfn main() {\n    let sin_usar = 3;\n    let mut x = 1;\n    println!(\"{}\", x);\n}\n",
}

RUNTIME = {
    "panic_index": "fn main() {\n    let v = vec![1, 2, 3];\n    let i = v.len() + 2;\n    println!(\"{}\", v[i]);\n}\n",
    "panic_unwrap_none": "fn main() {\n    let v: Vec<i32> = Vec::new();\n    let x = v.first().unwrap();\n    println!(\"{}\", x);\n}\n",
    "panic_unwrap_err": "fn main() {\n    let n: i32 = \"cinco\".parse().unwrap();\n    println!(\"{}\", n);\n}\n",
    "panic_expect": "fn main() {\n    let n: i32 = \"cinco\".parse().expect(\"esperaba un número\");\n    println!(\"{}\", n);\n}\n",
    "panic_overflow": "fn main() {\n    let mut x: u8 = 250;\n    for _ in 0..10 {\n        x += 1;\n    }\n    println!(\"{}\", x);\n}\n",
    "panic_div_zero": "fn main() {\n    let a = 10;\n    let b = \"0\".parse::<i32>().unwrap();\n    println!(\"{}\", a / b);\n}\n",
    "panic_str_boundary": "fn main() {\n    let s = String::from(\"ñandú\");\n    println!(\"{}\", &s[0..1]);\n}\n",
    "panic_refcell": "use std::cell::RefCell;\nfn main() {\n    let c = RefCell::new(1);\n    let _a = c.borrow_mut();\n    let _b = c.borrow_mut();\n}\n",
    "panic_explicit": "fn main() {\n    panic!(\"algo salió mal\");\n}\n",
    "panic_unreachable": "fn main() {\n    let x = 3;\n    if x > 2 {\n        unreachable!();\n    }\n}\n",
    "stack_overflow": "fn f(n: u64) -> u64 {\n    let a = [n; 64];\n    f(n + 1) + a[0]\n}\nfn main() {\n    println!(\"{}\", f(0));\n}\n",
    "main_err": "fn main() -> Result<(), String> {\n    Err(String::from(\"archivo no encontrado\"))\n}\n",
    "panic_nested_unwrap": "fn leer(s: &str) -> i32 {\n    s.parse().unwrap()\n}\nfn main() {\n    println!(\"{}\", leer(\"x\"));\n}\n",
}

CLIPPY = {
    "needless_range_loop": "fn main() {\n    let v = vec![1, 2, 3];\n    for i in 0..v.len() {\n        println!(\"{}\", v[i]);\n    }\n}\n",
    "redundant_clone": "fn main() {\n    let s = String::from(\"a\");\n    let t = s.clone();\n    println!(\"{}\", t);\n}\n",
}


def environment(cargo_home, rustup_home):
    env = dict(os.environ)
    env.update({"CARGO_HOME": cargo_home, "RUSTUP_HOME": rustup_home, "RUST_BACKTRACE": "1",
                "CARGO_TERM_COLOR": "never", "NO_COLOR": "1", "CARGO_INCREMENTAL": "0"})
    env["PATH"] = os.path.join(cargo_home, "bin") + os.pathsep + env.get("PATH", "")
    return env


def header(folder, command, source, extra=""):
    lines = "".join(f"#   {line}\n" for line in source.splitlines())
    return f"# workingDir: {folder}\n# command: {command}\n{extra}# program:\n{lines}# output:\n"


def run(args, cwd, env):
    # Windows looks for the program in the parent's PATH: resolve it in env's.
    args = [shutil.which(args[0], path=env["PATH"]) or args[0], *args[1:]]
    done = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=300)
    return done.returncode, done.stdout + done.stderr


def write(out_dir, case, text):
    os.makedirs(out_dir, exist_ok=True)
    with open(os.path.join(out_dir, case + ".txt"), "w", encoding="utf-8", newline="\n") as f:
        f.write(text)


def student_folder():
    work = os.path.join(tempfile.mkdtemp(), "mis programas ñandú")
    os.makedirs(work)
    return work


def record_loose(env, out, wanted):
    work = student_folder()
    for case, source in {**COMPILE, **RUNTIME}.items():
        if wanted and case not in wanted:
            continue
        with open(os.path.join(work, "main.rs"), "w", encoding="utf-8") as f:
            f.write(source)
        for mode, extra in (("json", ["--error-format=json"]), ("text", [])):
            args = ["rustc", "--edition", "2024", "-g", *extra, "-o", "main.exe", "main.rs"]
            code, text = run(args, work, env)
            meta = f"# compileExit: {code}\n"
            if case in RUNTIME and code == 0 and mode == "text":
                runtime_code, runtime_text = run([os.path.join(work, "main.exe")], work, env)
                write(os.path.join(out, "runtime"), case, header(work, "main.exe", source,
                      f"# runExit: {runtime_code} (0x{runtime_code & 0xFFFFFFFF:08X})\n") + runtime_text)
            if case in COMPILE:
                write(os.path.join(out, mode), case, header(work, " ".join(args), source, meta) + text)


def record_clippy(env, out, wanted):
    work = student_folder()
    for case, source in CLIPPY.items():
        if wanted and case not in wanted:
            continue
        with open(os.path.join(work, "main.rs"), "w", encoding="utf-8") as f:
            f.write(source)
        args = ["clippy-driver", "--edition", "2024", "--crate-type", "bin", "--emit=metadata",
                "--out-dir", "build", "--error-format=json", "main.rs"]
        code, text = run(args, work, env)
        write(os.path.join(out, "clippy"), case, header(work, " ".join(args), source, f"# exit: {code}\n") + text)


def record_cargo(env, out, wanted):
    """A project with a compile error, one that panics in a module and a workspace member."""
    work = student_folder()
    project = os.path.join(work, "juego")
    shutil.rmtree(project, ignore_errors=True)
    run(["cargo", "new", "--vcs", "none", "juego"], work, env)
    cases = {
        "build_error": ("src/main.rs", COMPILE["moved"], ["cargo", "build", "--message-format=json", "--quiet"]),
        "run_panic": ("src/main.rs", RUNTIME["panic_index"], ["cargo", "run", "--quiet"]),
        "clippy_project": ("src/main.rs", CLIPPY["needless_range_loop"], ["cargo", "clippy", "--message-format=json", "--quiet"]),
    }
    for case, (name, source, args) in cases.items():
        if wanted and case not in wanted:
            continue
        with open(os.path.join(project, name), "w", encoding="utf-8") as f:
            f.write(source)
        code, text = run(args, project, env)
        write(os.path.join(out, "cargo"), case, header(project, " ".join(args), source, f"# exit: {code}\n") + text)


def main():
    cargo_home, rustup_home, out = sys.argv[1:4]
    wanted = set(sys.argv[4:])
    env = environment(cargo_home, rustup_home)
    version = run(["rustc", "--version"], ".", env)[1].strip()
    record_loose(env, out, wanted)
    record_clippy(env, out, wanted)
    record_cargo(env, out, wanted)
    write(out, "VERSION", version + "\n")
    print(f"{version}: fixtures in {out}")


main()
