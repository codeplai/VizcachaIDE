// The text a new file starts with, per programming language.
import type { CodeLanguage } from '../domain'

const GO_HELLO = `package main

import "fmt"

func main() {
    fmt.Println("Hola, Go")
}
`

const GO_BLANK = `package main

func main() {
}
`

const PYTHON_HELLO = `print("Hola, Python")
`

const CPP_HELLO = `#include <iostream>
using namespace std;

int main() {
    cout << "Hola, C++" << endl;
    return 0;
}
`

const CPP_BLANK = `int main() {
    return 0;
}
`

const RUST_HELLO = `fn main() {
    println!("Hola, Rust");
}
`

const RUST_BLANK = `fn main() {}
`

export interface NewFileTemplate {
  /** A program that prints a greeting (the first-run wizard's "Hello" choice). */
  hello: string
  /** The smallest program the language can run (File > New). */
  blank: string
  /** Extension of the untitled name; the first one the profile declares. */
  extension: string
}

export const newFileTemplates: Record<CodeLanguage, NewFileTemplate> = {
  go: { hello: GO_HELLO, blank: GO_BLANK, extension: '.go' },
  python: { hello: PYTHON_HELLO, blank: '', extension: '.py' },
  cpp: { hello: CPP_HELLO, blank: CPP_BLANK, extension: '.cpp' },
  rust: { hello: RUST_HELLO, blank: RUST_BLANK, extension: '.rs' }
}
