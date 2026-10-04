// Part of the domain.ts contract (v3): the language profiles. Import it through domain.ts.

export type CodeLanguage = 'go' | 'python' | 'cpp'

export type PackageAction = 'init' | 'add' | 'remove' | 'tidy' | 'list'

export interface IndentStyle {
  useTabs: boolean
  size: number
}

/** Which buttons and panels apply to a language. */
export interface Capabilities {
  build: boolean
  console: boolean
  format: boolean
  check: boolean
  /** Empty: the language has no package manager. */
  packageActions: PackageAction[]
  /** i18n key: 'debug.goroutines' or 'debug.threads'. */
  threadsLabel: string
}

export type ToolRole = 'runtime' | 'compiler' | 'debugAdapter' | 'languageServer' | 'formatter'

/** One tool a language needs and how to get it when it is missing. */
export interface ToolSpec {
  id: string
  role: ToolRole
  labelKey: string
  missingKey: string
  installUrl: string
  installCommand: string
  /** Id of the tool that contains this one ('python' for debugpy); '' for an executable. */
  providedBy: string
}

export interface LanguageProfile {
  id: CodeLanguage
  /** i18n key: 'codeLanguage.go'... */
  nameKey: string
  /** Lower case, with the dot. */
  extensions: string[]
  indent: IndentStyle
  capabilities: Capabilities
  tools: ToolSpec[]
}

export type ToolSource = 'configured' | 'bundled' | 'path' | 'missing'

/** What the IDE found for one tool. */
export interface ToolStatus {
  id: string
  codeLanguage: CodeLanguage
  role: ToolRole
  /** '' when missing. */
  version: string
  source: ToolSource
  path: string
}
