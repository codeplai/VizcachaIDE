import type {
  CompletionItem,
  DocumentSymbol,
  InlayHint,
  SignatureHelp,
  SourceLocation,
  SourceRange
} from '../domain'
import type { RefactorApi } from './typesRefactor'

/** Mirrors bridge.LanguageService (Go): code intelligence from the language server of a file. */
export interface LanguageApi extends RefactorApi {
  openDocument: (path: string, text: string) => Promise<void>
  changeDocument: (path: string, text: string, version: number) => Promise<void>
  closeDocument: (path: string) => Promise<void>
  completion: (at: SourceLocation) => Promise<CompletionItem[]>
  hover: (at: SourceLocation) => Promise<string>
  definition: (at: SourceLocation) => Promise<SourceLocation | null>
  signatureHelp: (at: SourceLocation) => Promise<SignatureHelp | null>
  documentHighlights: (at: SourceLocation) => Promise<SourceRange[]>
  documentSymbols: (path: string) => Promise<DocumentSymbol[]>
  /** The hints of the visible lines of an open file (`visible.start.file`). */
  inlayHints: (visible: SourceRange) => Promise<InlayHint[]>
  /** Where a new unsaved file named `name` lives while open (an absolute temporary path). */
  untitledFile: (name: string) => Promise<string>
}
