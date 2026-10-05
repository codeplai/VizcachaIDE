// The language profiles as the backend declares them (adapters/golang/profile.go, python, cpp and
// rust).
// The mock serves them; the Wails bridge uses them only until bridge.CodeLanguagesService exists
// (track N5), then this file serves the mock alone.
import type { LanguageProfile } from '../domain'
import { rustProfile } from './languageProfileRust'
import { tool } from './toolSpec'

export { rustProfile }

export const goProfile: LanguageProfile = {
  id: 'go',
  nameKey: 'codeLanguage.go',
  extensions: ['.go'],
  indent: { useTabs: true, size: 4 },
  capabilities: {
    build: true,
    console: true,
    format: true,
    check: true,
    debugInput: false,
    packageActions: ['init', 'add', 'tidy'],
    threadsLabel: 'debug.goroutines'
  },
  tools: [
    tool({
      id: 'go',
      role: 'runtime',
      labelKey: 'settings.toolGo',
      missingKey: 'errors.goNotFound',
      installUrl: 'https://go.dev/dl/'
    }),
    tool({
      id: 'dlv',
      role: 'debugAdapter',
      labelKey: 'settings.toolDelve',
      missingKey: 'errors.delveNotFound',
      installCommand: 'go install github.com/go-delve/delve/cmd/dlv@latest'
    }),
    tool({
      id: 'gopls',
      role: 'languageServer',
      labelKey: 'settings.toolGopls',
      missingKey: 'errors.goplsNotFound',
      installCommand: 'go install golang.org/x/tools/gopls@latest'
    })
  ]
}

export const pythonProfile: LanguageProfile = {
  id: 'python',
  nameKey: 'codeLanguage.python',
  extensions: ['.py', '.pyw'],
  indent: { useTabs: false, size: 4 },
  capabilities: {
    build: false,
    console: true,
    format: true,
    check: true,
    debugInput: true,
    packageActions: ['add', 'remove', 'list'],
    threadsLabel: 'debug.threads'
  },
  tools: [
    tool({
      id: 'python',
      role: 'runtime',
      labelKey: 'settings.toolPython',
      missingKey: 'errors.pythonNotFound',
      installUrl: 'https://www.python.org/downloads/'
    }),
    tool({
      id: 'debugpy',
      role: 'debugAdapter',
      providedBy: 'python',
      labelKey: 'settings.modulePython',
      missingKey: 'errors.debugpyMissing',
      installCommand: 'python -m pip install debugpy'
    }),
    tool({
      id: 'pylsp',
      role: 'languageServer',
      providedBy: 'python',
      labelKey: 'settings.modulePython',
      missingKey: 'errors.pylspMissing',
      installCommand: 'python -m pip install "python-lsp-server[pyflakes]"'
    }),
    tool({
      id: 'ruff',
      role: 'formatter',
      providedBy: 'python',
      labelKey: 'settings.modulePython',
      missingKey: 'errors.ruffMissing',
      installCommand: 'python -m pip install ruff'
    })
  ]
}

export const cppProfile: LanguageProfile = {
  id: 'cpp',
  nameKey: 'codeLanguage.cpp',
  extensions: ['.cpp', '.cc', '.cxx', '.c++', '.h', '.hpp', '.hh'],
  indent: { useTabs: false, size: 4 },
  capabilities: {
    build: true,
    console: false,
    format: true,
    check: true,
    debugInput: true,
    packageActions: ['add', 'remove', 'list'],
    threadsLabel: 'debug.threads'
  },
  tools: [
    tool({
      id: 'cxx',
      role: 'compiler',
      labelKey: 'settings.toolCxx',
      missingKey: 'errors.cxxNotFound',
      installUrl: 'https://winlibs.com/'
    }),
    tool({
      id: 'lldb-dap',
      role: 'debugAdapter',
      labelKey: 'settings.toolLldbDap',
      missingKey: 'errors.lldbDapNotFound',
      installUrl: 'https://github.com/mstorsjo/llvm-mingw/releases'
    }),
    tool({
      id: 'clangd',
      role: 'languageServer',
      labelKey: 'settings.toolClangd',
      missingKey: 'errors.clangdNotFound',
      installUrl: 'https://clangd.llvm.org/installation'
    }),
    tool({
      id: 'clang-format',
      role: 'formatter',
      labelKey: 'settings.toolClangFormat',
      missingKey: 'errors.clangFormatNotFound',
      installUrl: 'https://github.com/mstorsjo/llvm-mingw/releases'
    }),
    tool({
      id: 'cmake',
      role: 'buildTool',
      labelKey: 'settings.toolCMake',
      missingKey: 'errors.cmakeNotFound',
      installUrl: 'https://cmake.org/download/'
    }),
    tool({
      id: 'ninja',
      role: 'buildTool',
      labelKey: 'settings.toolNinja',
      missingKey: 'errors.ninjaNotFound',
      installUrl: 'https://github.com/ninja-build/ninja/releases'
    }),
    tool({
      id: 'vcpkg',
      role: 'buildTool',
      labelKey: 'settings.toolVcpkg',
      missingKey: 'errors.vcpkgNotFound',
      installUrl: 'https://github.com/microsoft/vcpkg'
    })
  ]
}

export const languageProfiles: LanguageProfile[] = [
  goProfile,
  pythonProfile,
  cppProfile,
  rustProfile
]
