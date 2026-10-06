import type { SnippetDef } from './types'

export const rustSnippets: SnippetDef[] = [
  {
    label: 'for',
    template: 'for ${item} in ${items} {\n\t${}\n}',
    description: 'snippets.forEach'
  },
  { label: 'if', template: 'if ${condition} {\n\t${}\n}', description: 'snippets.if' },
  { label: 'while', template: 'while ${condition} {\n\t${}\n}', description: 'snippets.while' },
  { label: 'fn', template: 'fn ${name}(${}) ${} {\n\t${}\n}', description: 'snippets.func' },
  { label: 'main', template: 'fn main() {\n\t${}\n}', description: 'snippets.main' },
  { label: 'println', template: 'println!("{}", ${});', description: 'snippets.printRust' },
  {
    label: 'struct',
    template: 'struct ${Name} {\n\t${field}: ${String},\n}',
    description: 'snippets.struct'
  },
  {
    label: 'match',
    template: 'match ${value} {\n\t${pattern} => ${},\n\t_ => ${},\n}',
    description: 'snippets.match'
  }
]
