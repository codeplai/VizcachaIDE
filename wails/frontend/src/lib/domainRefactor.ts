// Part of the domain.ts contract: renaming symbols and finding references. Import it through domain.ts.

import type { SourceRange } from './domain'

/** Why a symbol cannot be renamed (the editor translates it). Empty: it can. */
export type RenameRefusal = '' | 'notRenameable' | 'unsupported' | 'failed'

/** The answer of prepareRename: the symbol to rename, when the server says which one. */
export interface RenameTarget {
  refusal: RenameRefusal
  range?: SourceRange
  placeholder: string
  detail?: string
}

/** Replaces a range (1-based lines, columns in runes, end exclusive) with `newText`. */
export interface TextEdit {
  range: SourceRange
  newText: string
}

export interface FileEdit {
  file: string
  edits: TextEdit[]
}

/** The edits of a rename over every affected file, or why there are none. */
export interface RenameResult {
  refusal: RenameRefusal
  detail?: string
  files: FileEdit[]
}

/** One use of a symbol, with the text of its line. */
export interface Reference {
  range: SourceRange
  preview: string
}

export interface EditSummary {
  files: number
  edits: number
}
