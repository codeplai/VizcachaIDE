// The CodeMirror extensions that depend on a file's programming language.
import { cpp } from '@codemirror/lang-cpp'
import { go } from '@codemirror/lang-go'
import { python } from '@codemirror/lang-python'
import { rust } from '@codemirror/lang-rust'
import { indentUnit } from '@codemirror/language'
import { EditorState, type Extension } from '@codemirror/state'
import type { CodeLanguage, IndentStyle, LanguageProfile } from '../domain'

/** Go's indentation when no profile is known (before they load, or a file of no language). */
const GO_INDENT: IndentStyle = { useTabs: true, size: 4 }

/** Syntax highlighting per language (the cpp profile owns .h and .hpp too). */
const syntaxFor = (codeLanguage: CodeLanguage | undefined): Extension => {
  if (codeLanguage === 'go') return go()
  if (codeLanguage === 'python') return python()
  if (codeLanguage === 'cpp') return cpp()
  if (codeLanguage === 'rust') return rust()
  return []
}

const indentation = (indent: IndentStyle): Extension => [
  indentUnit.of(indent.useTabs ? '\t' : ' '.repeat(indent.size)),
  EditorState.tabSize.of(indent.size)
]

const PLAIN_INDENT: IndentStyle = { useTabs: false, size: 4 }

/**
 * The language extensions of one file: its syntax plus the indentation its profile asks for.
 * With no profile (still loading, or a file no language claims) a .go file keeps Go's settings and
 * anything else is plain text.
 */
export const languageExtensionsFor = (
  path: string,
  profile: LanguageProfile | null
): Extension[] => {
  const goFile = path.toLowerCase().endsWith('.go')
  const codeLanguage = profile?.id ?? (goFile ? 'go' : undefined)
  const indent = profile?.indent ?? (goFile ? GO_INDENT : PLAIN_INDENT)
  return [syntaxFor(codeLanguage), indentation(indent)]
}
