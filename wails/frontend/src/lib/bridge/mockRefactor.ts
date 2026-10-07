// Rename and find references for the mock bridge: a whole-word search over the files of the demo
// project, standing in for the language server so `npm run dev` shows the real flow.
import type {
  EditSummary,
  FileEdit,
  FileNode,
  Reference,
  RenameResult,
  RenameTarget,
  SourceLocation,
  SourceRange
} from '../domain'
import { applyChanges, editsToChanges, runesOfUnits, unitsOfRunes } from '../textEdits'
import type { FilesApi, LanguageApi } from './types'

/** Words that are never a name to rename (Go, Python, C++ and Rust keywords of the demo files). */
const KEYWORDS = new Set(
  (
    'break case chan const continue default defer else fallthrough for func go goto if import ' +
    'interface map package range return select struct switch type var def class from while ' +
    'with yield lambda pass in is not and or None True False fn let mut impl use pub mod match ' +
    'loop self int string bool float void auto'
  ).split(' ')
)

const WORD = /[\p{L}\p{N}_]+/gu
const IDENTIFIER = /^[\p{L}_][\p{L}\p{N}_]*$/u

/** The word at a 1-based line and rune column, with its UTF-16 offsets in the line. */
const wordAt = (text: string, at: SourceLocation): { word: string; start: number } | null => {
  const line = text.split('\n')[at.line - 1] ?? ''
  const unit = unitsOfRunes(line, at.column - 1)
  for (const match of line.matchAll(WORD)) {
    const start = match.index
    if (unit >= start && unit <= start + match[0].length) return { word: match[0], start }
  }
  return null
}

const flatten = (node: FileNode): string[] =>
  node.isDir ? node.children.flatMap(flatten) : [node.path]

const extensionOf = (path: string): string => path.slice(path.lastIndexOf('.'))

type Documents = Map<string, string>

/** Where the word appears in the project: the file, the 1-based line and the rune column. */
const occurrences = async (
  files: FilesApi,
  documents: Documents,
  word: string,
  like: string
): Promise<{ file: string; line: number; column: number; text: string }[]> => {
  const found: { file: string; line: number; column: number; text: string }[] = []
  const paths = flatten(await files.openFolder()).filter(
    (p) => extensionOf(p) === extensionOf(like)
  )
  for (const file of paths) {
    const text = documents.get(file) ?? (await files.readFile(file))
    text.split('\n').forEach((line, index) => {
      for (const match of line.matchAll(WORD)) {
        if (match[0] !== word) continue
        const column = runesOfUnits(line, match.index) + 1
        found.push({ file, line: index + 1, column, text: line })
      }
    })
  }
  return found
}

const rangeAt = (file: string, line: number, column: number, length: number): SourceRange => ({
  start: { file, line, column },
  end: { file, line, column: column + length }
})

const lengthOf = (word: string): number => runesOfUnits(word, word.length)

export const mockRefactor = (
  files: FilesApi,
  documents: Documents
): Pick<LanguageApi, 'prepareRename' | 'rename' | 'references'> => {
  const textOf = async (file: string): Promise<string> =>
    documents.get(file) ?? (await files.readFile(file))
  const prepareRename = async (at: SourceLocation): Promise<RenameTarget> => {
    const text = await textOf(at.file)
    const hit = wordAt(text, at)
    if (!hit || KEYWORDS.has(hit.word)) return { refusal: 'notRenameable', placeholder: '' }
    const line = text.split('\n')[at.line - 1] ?? ''
    const column = runesOfUnits(line, hit.start) + 1
    return {
      refusal: '',
      placeholder: hit.word,
      range: rangeAt(at.file, at.line, column, lengthOf(hit.word))
    }
  }
  const rename = async (at: SourceLocation, newName: string): Promise<RenameResult> => {
    const target = await prepareRename(at)
    if (target.refusal) return { refusal: 'notRenameable', files: [] }
    if (!IDENTIFIER.test(newName)) {
      return { refusal: 'failed', detail: `“${newName}” is not a valid name`, files: [] }
    }
    const byFile = new Map<string, FileEdit>()
    for (const hit of await occurrences(files, documents, target.placeholder, at.file)) {
      const edit = byFile.get(hit.file) ?? { file: hit.file, edits: [] }
      const range = rangeAt(hit.file, hit.line, hit.column, lengthOf(target.placeholder))
      edit.edits.push({ range, newText: newName })
      byFile.set(hit.file, edit)
    }
    return { refusal: '', files: [...byFile.values()] }
  }
  const references = async (at: SourceLocation): Promise<Reference[]> => {
    const target = await prepareRename(at)
    if (target.refusal) return []
    const hits = await occurrences(files, documents, target.placeholder, at.file)
    return hits.map((hit) => ({
      range: rangeAt(hit.file, hit.line, hit.column, lengthOf(target.placeholder)),
      preview: hit.text.trim()
    }))
  }
  return { prepareRename, rename, references }
}

/** Applies edits to the mock's in-memory files. */
export const applyMockEdits = (texts: Record<string, string>, files: FileEdit[]): EditSummary => {
  let edits = 0
  for (const file of files) {
    const text = texts[file.file] ?? ''
    const changes = editsToChanges(text, file.edits)
    texts[file.file] = applyChanges(text, changes)
    edits += file.edits.length
  }
  return { files: files.length, edits }
}
