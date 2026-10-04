// The Rust profile as the backend declares it (adapters/rust/profile.go).
import type { LanguageProfile } from '../domain'
import { tool } from './toolSpec'

export const rustProfile: LanguageProfile = {
  id: 'rust',
  nameKey: 'codeLanguage.rust',
  extensions: ['.rs'],
  indent: { useTabs: false, size: 4 },
  capabilities: {
    build: true,
    console: false,
    format: true,
    check: true,
    debugInput: true,
    packageActions: ['init', 'add', 'remove', 'list'],
    threadsLabel: 'debug.threads'
  },
  tools: [
    tool({
      id: 'rustc',
      role: 'compiler',
      labelKey: 'settings.toolRustc',
      missingKey: 'errors.rustNotFound',
      installUrl: 'https://rustup.rs'
    }),
    tool({
      id: 'cargo',
      role: 'runtime',
      labelKey: 'settings.toolCargo',
      missingKey: 'errors.cargoNotFound',
      installUrl: 'https://rustup.rs'
    }),
    tool({
      id: 'rust-analyzer',
      role: 'languageServer',
      labelKey: 'settings.toolRustAnalyzer',
      missingKey: 'errors.rustAnalyzerNotFound',
      installCommand: 'rustup component add rust-analyzer'
    }),
    tool({
      id: 'lldb-dap',
      role: 'debugAdapter',
      labelKey: 'settings.toolLldbDap',
      missingKey: 'errors.lldbDapNotFound',
      installUrl: 'https://github.com/mstorsjo/llvm-mingw/releases'
    }),
    // Components of the toolchain: a status row, no "Choose" button.
    tool({
      id: 'clippy',
      role: 'compiler',
      providedBy: 'cargo',
      labelKey: 'settings.toolClippy',
      missingKey: 'errors.clippyNotFound',
      installCommand: 'rustup component add clippy'
    }),
    tool({
      id: 'rustfmt',
      role: 'formatter',
      providedBy: 'cargo',
      labelKey: 'settings.toolRustfmt',
      missingKey: 'errors.rustfmtNotFound',
      installCommand: 'rustup component add rustfmt'
    })
  ]
}
