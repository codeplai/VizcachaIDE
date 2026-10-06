import type { Reference, RenameResult, RenameTarget, SourceLocation } from '../domain'

/** Rename and find references, part of LanguageApi (bridge.LanguageService in Go). */
export interface RefactorApi {
  /** Whether the symbol at a position can be renamed, and which text is its name. */
  prepareRename: (at: SourceLocation) => Promise<RenameTarget>
  /** The edits that rename the symbol, over every file it touches. Nothing is written. */
  rename: (at: SourceLocation, newName: string) => Promise<RenameResult>
  /** Every use of the symbol at a position, its declaration included. */
  references: (at: SourceLocation) => Promise<Reference[]>
}
