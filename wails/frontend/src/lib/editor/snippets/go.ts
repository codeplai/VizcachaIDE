import type { SnippetDef } from './types'

export const goSnippets: SnippetDef[] = [
  {
    label: 'for',
    template: 'for ${i} := 0; ${i} < ${n}; ${i}++ {\n\t${}\n}',
    description: 'snippets.forCount'
  },
  {
    label: 'range',
    template: 'for ${i}, ${item} := range ${items} {\n\t${}\n}',
    description: 'snippets.forEach'
  },
  { label: 'if', template: 'if ${condition} {\n\t${}\n}', description: 'snippets.if' },
  { label: 'while', template: 'for ${condition} {\n\t${}\n}', description: 'snippets.whileGo' },
  {
    label: 'func',
    template: 'func ${name}(${}) ${} {\n\t${}\n}',
    description: 'snippets.func'
  },
  { label: 'main', template: 'func main() {\n\t${}\n}', description: 'snippets.main' },
  { label: 'println', template: 'fmt.Println(${})', description: 'snippets.printGo' },
  {
    label: 'struct',
    template: 'type ${Name} struct {\n\t${Field} ${string}\n}',
    description: 'snippets.struct'
  },
  {
    label: 'switch',
    template: 'switch ${value} {\ncase ${1}:\n\t${}\ndefault:\n\t${}\n}',
    description: 'snippets.switch'
  }
]
