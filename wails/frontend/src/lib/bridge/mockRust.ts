// Demo data of the Rust sample project (`?language=rust` in the browser): a Cargo project with the
// states of the other languages plus a panic, with a factorial(n: u64) to look at while debugging.
// The errors and their explanations are in mockRustErrors.ts.
import type {
  DebugState,
  FileNode,
  FrameVariables,
  SourceLocation,
  StackFrame,
  Variable
} from '../domain'

export const RUST_DIR = 'hola-rust'
export const RUST_MANIFEST = 'hola-rust/Cargo.toml'
export const RUST_MAIN = 'hola-rust/src/main.rs'
/** The line the dev bar puts a breakpoint on: the base case of factorial. */
export const RUST_BREAKPOINT_LINE = 5
export const RUST_LAST_LINE = 15
/** The manifest of a virtual workspace: running it asks which member to run (`run.chooseMember`). */
export const RUST_WORKSPACE_MANIFEST = 'taller-rust/Cargo.toml'
export const RUST_WORKSPACE_ERROR = 'run.chooseMember: app, cli'
/** What the build stage prints before the program runs (the backend's own line). */
export const RUST_COMPILING_NOTICE = 'Compiling hola-rust v0.1.0'

export const RUST_SOURCES: Record<string, string> = {
  [RUST_MANIFEST]: `[package]
name = "hola-rust"
version = "0.1.0"
edition = "2024"

[dependencies]
`,
  [RUST_MAIN]: `use std::io;

fn factorial(n: u64) -> u64 {
    if n <= 1 {
        return 1;
    }
    n * factorial(n - 1)
}

fn main() {
    let mut nombre = String::new();
    io::stdin().read_line(&mut nombre).unwrap();
    let resultado = factorial(3);
    println!("Hola, Rust");
}
`
}

export const rustTree = (): FileNode => ({
  name: RUST_DIR,
  path: RUST_DIR,
  isDir: true,
  children: [
    { name: 'Cargo.toml', path: RUST_MANIFEST, isDir: false, children: [] },
    {
      name: 'src',
      path: `${RUST_DIR}/src`,
      isDir: true,
      children: [{ name: 'main.rs', path: RUST_MAIN, isDir: false, children: [] }]
    }
  ]
})

const at = (line: number): SourceLocation => ({ file: RUST_MAIN, line, column: 1 })

const variable = (name: string, typeName: string, value: string, changed = false): Variable => ({
  name,
  typeName,
  value,
  reference: 0,
  changed,
  children: []
})

const FACTORIAL = 'hola_rust::factorial'

/** factorial(3) paused at its base case: three calls deep plus main. */
const CALLERS: StackFrame[] = [
  { frameId: 2, function: FACTORIAL, location: at(7) },
  { frameId: 3, function: FACTORIAL, location: at(7) },
  { frameId: 4, function: 'hola_rust::main', location: at(13) }
]

export const rustDebugState = (line = RUST_BREAKPOINT_LINE): DebugState => ({
  reason: line === RUST_BREAKPOINT_LINE ? 'breakpoint' : 'step',
  frames: [{ frameId: 1, function: FACTORIAL, location: at(line) }, ...CALLERS],
  variables: [variable('n', 'u64', '1', line === RUST_BREAKPOINT_LINE)],
  threads: [{ threadId: 1, name: 'main', location: at(line) }],
  currentThread: 1,
  description: ''
})

/** Arguments and locals of each frame of the sample stack. */
export const rustFrameVariables = (frameId: number): FrameVariables => {
  const byFrame: Record<number, FrameVariables> = {
    1: { arguments: [variable('n', 'u64', '1')], locals: [] },
    2: { arguments: [variable('n', 'u64', '2')], locals: [] },
    3: { arguments: [variable('n', 'u64', '3')], locals: [] },
    4: {
      arguments: [],
      locals: [variable('nombre', 'String', '"Ana"'), variable('resultado', 'u64', '0')]
    }
  }
  return byFrame[frameId] ?? { arguments: [], locals: [] }
}
