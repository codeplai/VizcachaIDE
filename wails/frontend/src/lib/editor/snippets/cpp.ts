import type { SnippetDef } from './types'

export const cppSnippets: SnippetDef[] = [
  {
    label: 'for',
    template: 'for (int ${i} = 0; ${i} < ${n}; ${i}++) {\n\t${}\n}',
    description: 'snippets.forCount'
  },
  {
    label: 'foreach',
    template: 'for (auto& ${item} : ${items}) {\n\t${}\n}',
    description: 'snippets.forEach'
  },
  { label: 'if', template: 'if (${condition}) {\n\t${}\n}', description: 'snippets.if' },
  { label: 'while', template: 'while (${condition}) {\n\t${}\n}', description: 'snippets.while' },
  {
    label: 'func',
    template: '${int} ${name}(${}) {\n\t${}\n}',
    description: 'snippets.func'
  },
  {
    label: 'main',
    template: 'int main() {\n\t${}\n\treturn 0;\n}',
    description: 'snippets.main'
  },
  {
    label: 'cout',
    template: 'std::cout << ${} << std::endl;',
    description: 'snippets.printCpp'
  },
  {
    label: 'class',
    template: 'class ${Name} {\npublic:\n\t${Name}(${}) {}\n\nprivate:\n\t${}\n};',
    description: 'snippets.class'
  },
  {
    label: 'struct',
    template: 'struct ${Name} {\n\t${int} ${field};\n};',
    description: 'snippets.struct'
  },
  {
    label: 'switch',
    template: 'switch (${value}) {\ncase ${1}:\n\t${}\n\tbreak;\ndefault:\n\tbreak;\n}',
    description: 'snippets.switch'
  },
  {
    label: 'try',
    template: 'try {\n\t${}\n} catch (const std::exception& ${e}) {\n\t${}\n}',
    description: 'snippets.try'
  }
]
