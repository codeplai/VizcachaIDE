import type { SnippetDef } from './types'

export const pythonSnippets: SnippetDef[] = [
  {
    label: 'for',
    template: 'for ${item} in ${items}:\n\t${}',
    description: 'snippets.forEach'
  },
  { label: 'if', template: 'if ${condition}:\n\t${}', description: 'snippets.if' },
  { label: 'while', template: 'while ${condition}:\n\t${}', description: 'snippets.while' },
  { label: 'def', template: 'def ${name}(${}):\n\t${}', description: 'snippets.func' },
  {
    label: 'main',
    template: 'if __name__ == "__main__":\n\t${}',
    description: 'snippets.main'
  },
  { label: 'print', template: 'print(${})', description: 'snippets.printPython' },
  {
    label: 'class',
    template: 'class ${Name}:\n\tdef __init__(self${}):\n\t\t${}',
    description: 'snippets.class'
  },
  {
    label: 'try',
    template: 'try:\n\t${}\nexcept ${Exception} as ${error}:\n\t${}',
    description: 'snippets.try'
  },
  {
    label: 'match',
    template: 'match ${value}:\n\tcase ${pattern}:\n\t\t${}\n\tcase _:\n\t\t${}',
    description: 'snippets.match'
  }
]
