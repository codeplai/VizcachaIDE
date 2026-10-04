// The errors of the Rust sample (mockRust.ts): a move rustc rejects (E0382) and a panic at run time.
import type { Diagnostic, ErrorExplanation, ExplainedDiagnostic } from '../domain'
import { RUST_MAIN } from './mockRust'

type Texts = Record<'en' | 'es', ErrorExplanation>

/** The moved variable of the error scenario: rustc's rendered text, the way the runner prints it. */
export const rustDiagnostic = (): Diagnostic => ({
  location: { file: RUST_MAIN, line: 14, column: 26 },
  end: { file: RUST_MAIN, line: 14, column: 32 },
  severity: 'error',
  message: 'borrow of moved value: `nombre`',
  rawText: [
    'error[E0382]: borrow of moved value: `nombre`',
    '  --> src\\main.rs:14:26',
    '   |',
    '11 |     let mut nombre = String::new();',
    '   |         ---------- move occurs because `nombre` has type `String`, which does not implement the `Copy` trait',
    '12 |     let copia = nombre;',
    '   |                 ------ value moved here',
    '13 |',
    '14 |     println!("Hola, {}", nombre);',
    '   |                          ^^^^^^ value borrowed here after move'
  ].join('\n'),
  source: 'compiler',
  code: 'E0382'
})

/** The panic of the crash scenario: what the runtime prints, with the frame of the student's code. */
export const rustPanicDiagnostic = (): Diagnostic => ({
  location: { file: RUST_MAIN, line: 13, column: 20 },
  end: null,
  severity: 'error',
  message: 'index out of bounds: the len is 3 but the index is 5',
  rawText: [
    "thread 'main' (35236) panicked at src\\main.rs:13:20:",
    'index out of bounds: the len is 3 but the index is 5',
    'stack backtrace:',
    '   6: hola_rust::main',
    '             at .\\src\\main.rs:13:20'
  ].join('\n'),
  source: 'runtime',
  code: 'RS-PANIC-INDEX'
})

const MOVED: Texts = {
  en: {
    explanationId: 'RS-MOVED',
    title: 'The value of “nombre” was moved and can’t be used again',
    body: 'On line 12 the value of “nombre” was handed over to another variable: in Rust a String has one owner at a time, so after that move the original name is empty. On line 14 you tried to use it again.',
    fixHint:
      'Lend it instead of moving it (let copia = &nombre;) or make a real copy (let copia = nombre.clone();).',
    placeholders: { name: 'nombre', line: '14' }
  },
  es: {
    explanationId: 'RS-MOVED',
    title: 'El valor de «nombre» se movió y no se puede usar otra vez',
    body: 'En la línea 12 el valor de «nombre» pasó a otra variable: en Rust un String tiene un solo dueño a la vez, así que después de ese movimiento el nombre original queda vacío. En la línea 14 intentaste usarlo de nuevo.',
    fixHint:
      'Préstalo en vez de moverlo (let copia = &nombre;) o haz una copia de verdad (let copia = nombre.clone();).',
    placeholders: { name: 'nombre', line: '14' }
  }
}

const PANIC_INDEX: Texts = {
  en: {
    explanationId: 'RS-PANIC-INDEX',
    title: 'The program asked for a position that doesn’t exist',
    body: 'The list has 3 items (positions 0, 1 and 2) but the program asked for position 5. Rust stops the program with a panic instead of reading memory that isn’t part of the list.',
    fixHint:
      'Check the index before using it (v.get(5) answers None when it is missing) or fix the number.',
    placeholders: { len: '3', index: '5' }
  },
  es: {
    explanationId: 'RS-PANIC-INDEX',
    title: 'El programa pidió una posición que no existe',
    body: 'La lista tiene 3 elementos (posiciones 0, 1 y 2) pero el programa pidió la posición 5. Rust detiene el programa con un panic en vez de leer memoria que no es de la lista.',
    fixHint:
      'Revisa el índice antes de usarlo (v.get(5) responde None si no existe) o corrige el número.',
    placeholders: { len: '3', index: '5' }
  }
}

export const rustExplanation = (language: 'en' | 'es'): ErrorExplanation => MOVED[language]

export const rustCrashExplained = (language: 'en' | 'es'): ExplainedDiagnostic[] => [
  { diagnostic: rustPanicDiagnostic(), explanation: PANIC_INDEX[language] }
]
