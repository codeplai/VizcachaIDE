// The projects of the multi-file phase (steps-multifile.mjs): in every language main calls a
// function of another file, which calls a function of a third one (impuesto), and a fourth file
// uses an external library. Same numbers everywhere: total([100, 250, 50]) = 400 + 10% = 440.

const go = {
  dir: 'go',
  files: {
    'go.mod': 'module example.com/tienda\n\ngo 1.21\n',
    'main.go':
      'package main\n\nimport (\n\t"fmt"\n\n\t"example.com/tienda/calculo"\n\t"example.com/tienda/texto"\n)\n\nfunc main() {\n\ttotal := calculo.Total([]int{100, 250, 50})\n\tfmt.Println(texto.Moneda(total))\n\tfmt.Println(texto.Recibo())\n}\n',
    'calculo/calculo.go':
      'package calculo\n\n// Total suma los precios y aplica el impuesto.\nfunc Total(precios []int) int {\n\tsuma := 0\n\tfor _, p := range precios {\n\t\tsuma += p\n\t}\n\treturn conImpuesto(suma)\n}\n',
    'calculo/impuesto.go':
      'package calculo\n\nfunc conImpuesto(monto int) int {\n\timpuesto := monto / 10\n\treturn monto + impuesto\n}\n',
    'texto/texto.go':
      'package texto\n\nimport (\n\t"fmt"\n\n\t"github.com/google/uuid"\n)\n\nfunc Moneda(valor int) string {\n\treturn fmt.Sprintf("Total: $%d", valor)\n}\n\nfunc Recibo() string {\n\tif len(uuid.NewString()) == 36 {\n\t\treturn "Recibo listo"\n\t}\n\treturn "Recibo roto"\n}\n'
  },
  library: { codeLanguage: 'go', name: 'github.com/google/uuid', added: ['go.mod', 'google/uuid'] },
  main: 'main.go',
  output: ['Total: $440', 'Recibo listo'],
  secondary: { file: 'calculo/impuesto.go', from: 'monto / 10', to: 'monto / diez' },
  breakpoint: { file: 'calculo/impuesto.go', line: 5 },
  stack: ['impuesto.go', 'calculo.go', 'main.go'],
  definition: { line: 11, word: 'Total', files: ['calculo.go'] },
  external: { file: 'texto/texto.go', line: 14, word: 'NewString' }
}

const python = {
  dir: 'py',
  files: {
    'main.py': 'from calculo import total\nfrom texto import moneda, grande\n\nprint(moneda(total([100, 250, 50])))\nprint(grande(1234567))\n',
    'calculo.py':
      'from impuesto import con_impuesto\n\n\ndef total(precios):\n    suma = 0\n    for p in precios:\n        suma += p\n    return con_impuesto(suma)\n',
    'impuesto.py': 'def con_impuesto(monto):\n    impuesto = monto // 10\n    return monto + impuesto\n',
    'texto.py':
      'import humanize\n\n\ndef moneda(valor):\n    return f"Total: ${valor}"\n\n\ndef grande(n):\n    return "Grande: " + humanize.intcomma(n)\n'
  },
  library: { codeLanguage: 'python', name: 'humanize' },
  main: 'main.py',
  output: ['Total: $440', 'Grande: 1,234,567'],
  secondary: { file: 'impuesto.py', from: 'monto // 10', to: 'monto // diez' },
  breakpoint: { file: 'impuesto.py', line: 3 },
  stack: ['impuesto.py', 'calculo.py', 'main.py'],
  definition: { line: 4, word: 'total', files: ['calculo.py'] },
  external: { file: 'texto.py', line: 9, word: 'intcomma' }
}

const cpp = {
  dir: 'cpp',
  files: {
    'calculo.h': '#pragma once\n#include <vector>\n\nint total(const std::vector<int>& precios);\nint conImpuesto(int monto);\n',
    'calculo.cpp':
      '#include "calculo.h"\n\nint total(const std::vector<int>& precios) {\n    int suma = 0;\n    for (int p : precios) {\n        suma += p;\n    }\n    return conImpuesto(suma);\n}\n',
    'impuesto.cpp': '#include "calculo.h"\n\nint conImpuesto(int monto) {\n    int impuesto = monto / 10;\n    return monto + impuesto;\n}\n',
    'main.cpp':
      '#include <iostream>\n#include "calculo.h"\n#include "json.hpp"\n\nint main() {\n    std::cout << "Total: $" << total({100, 250, 50}) << std::endl;\n    nlohmann::json recibo = {{"total", total({100, 250, 50})}};\n    std::cout << "Recibo: " << recibo.dump() << std::endl;\n    return 0;\n}\n'
  },
  // C++ has no package manager in the IDE: a header-only library copied next to the code.
  library: { download: 'https://github.com/nlohmann/json/releases/download/v3.12.0/json.hpp', file: 'json.hpp' },
  main: 'main.cpp',
  output: ['Total: $440', 'Recibo: {"total":440}'],
  secondary: { file: 'impuesto.cpp', from: 'monto / 10', to: 'monto / diez' },
  breakpoint: { file: 'impuesto.cpp', line: 5 },
  stack: ['impuesto.cpp', 'calculo.cpp', 'main.cpp'],
  definition: { line: 6, word: 'total', files: ['calculo.h', 'calculo.cpp'] },
  external: { file: 'main.cpp', line: 8, word: 'dump' }
}

const rust = {
  dir: 'rs',
  files: {
    'Cargo.toml': '[package]\nname = "tienda"\nversion = "0.1.0"\nedition = "2024"\n\n[dependencies]\n',
    'src/main.rs':
      'mod calculo;\nmod impuesto;\nmod texto;\n\nfn main() {\n    let total = calculo::total(&[100, 250, 50]);\n    println!("{}", texto::moneda(total));\n    println!("{}", texto::dado());\n}\n',
    'src/calculo.rs':
      'use crate::impuesto::con_impuesto;\n\npub fn total(precios: &[i32]) -> i32 {\n    let mut suma = 0;\n    for p in precios {\n        suma += p;\n    }\n    con_impuesto(suma)\n}\n',
    'src/impuesto.rs': 'pub fn con_impuesto(monto: i32) -> i32 {\n    let impuesto = monto / 10;\n    monto + impuesto\n}\n',
    'src/texto.rs':
      'pub fn moneda(valor: i32) -> String {\n    format!("Total: ${}", valor)\n}\n\npub fn dado() -> String {\n    let valor = rand::random::<u8>() % 6 + 1;\n    if (1..=6).contains(&valor) {\n        String::from("Dado listo")\n    } else {\n        String::from("Dado roto")\n    }\n}\n'
  },
  library: { codeLanguage: 'rust', name: 'rand', added: ['Cargo.toml', 'rand ='] },
  main: 'src/main.rs',
  output: ['Total: $440', 'Dado listo'],
  secondary: { file: 'src/impuesto.rs', from: 'monto / 10', to: 'monto / diez' },
  breakpoint: { file: 'src/impuesto.rs', line: 3 },
  stack: ['impuesto.rs', 'calculo.rs', 'main.rs'],
  definition: { line: 6, word: 'total(&', files: ['calculo.rs'] },
  external: { file: 'src/texto.rs', line: 6, word: 'random' }
}

export const MULTIFILE_PROJECTS = { go, python, cpp, rust }
