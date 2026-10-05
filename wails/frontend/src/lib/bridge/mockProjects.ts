// Mock of the Projects service: creates the "Hello + keyboard input" template in the mock file
// system, so `npm run dev` shows the new project in the file tree. The real templates live in Go.
import type { CodeLanguage } from '../domain'
import type { FilesApi, ProjectsApi } from './types'

const GREETING = '¿Cómo te llamas? '

const TEMPLATES: Partial<
  Record<CodeLanguage, { main: string; files: (name: string) => Record<string, string> }>
> = {
  go: {
    main: 'main.go',
    files: (name) => ({
      'go.mod': `module ${name.toLowerCase().replace(/[^a-z0-9._-]+/g, '-')}\n\ngo 1.21\n`,
      'main.go': `package main\n\nimport (\n\t"bufio"\n\t"fmt"\n\t"os"\n\t"strings"\n)\n\nfunc main() {\n\tfmt.Print("${GREETING}")\n\tnombre, _ := bufio.NewReader(os.Stdin).ReadString('\\n')\n\tfmt.Printf("Hola, %s\\n", strings.TrimSpace(nombre))\n}\n`
    })
  },
  python: {
    main: 'main.py',
    files: () => ({ 'main.py': `nombre = input("${GREETING}")\nprint(f"Hola, {nombre}")\n` })
  },
  cpp: {
    main: 'main.cpp',
    files: () => ({
      'main.cpp': `#include <iostream>\n#include <string>\n\nint main() {\n    std::cout << "${GREETING}";\n    std::string nombre;\n    std::getline(std::cin, nombre);\n    std::cout << "Hola, " << nombre << std::endl;\n}\n`,
      'compile_flags.txt': '-std=c++17\n-Wall\n-Wextra\n',
      '.clang-format': 'BasedOnStyle: LLVM\nIndentWidth: 4\n'
    })
  },
  rust: {
    main: 'src/main.rs',
    files: (name) => ({
      'Cargo.toml': `[package]\nname = "${name.toLowerCase().replace(/[^a-z0-9-]+/g, '_')}"\nversion = "0.1.0"\nedition = "2021"\n\n[dependencies]\n`,
      'src/main.rs': `use std::io::{self, Write};\n\nfn main() {\n    print!("${GREETING}");\n    io::stdout().flush().unwrap();\n    let mut nombre = String::new();\n    io::stdin().read_line(&mut nombre).unwrap();\n    println!("Hola, {}", nombre.trim());\n}\n`,
      '.gitignore': '/target\n'
    })
  }
}

const FORBIDDEN = /[\\/:*?"<>|]/

export const mockProjects = (files: FilesApi): ProjectsApi => ({
  create: async (codeLanguage, location, rawName) => {
    const name = rawName.trim()
    if (!name) throw new Error('project.errorNameEmpty')
    if (FORBIDDEN.test(name) || /^[. ]*$/.test(name)) throw new Error('project.errorNameInvalid')
    const template = TEMPLATES[codeLanguage]
    if (!template) throw new Error('project.errorNoScaffold')
    const root = `${location}/${name}`
    try {
      await files.createFolder(root)
    } catch {
      throw new Error('project.errorExists')
    }
    for (const [relative, text] of Object.entries(template.files(name))) {
      const parts = relative.split('/')
      if (parts.length > 1)
        await files.createFolder(`${root}/${parts.slice(0, -1).join('/')}`).catch(() => {})
      await files.createFile(`${root}/${relative}`, text)
    }
    return { root, mainFile: `${root}/${template.main}` }
  }
})
