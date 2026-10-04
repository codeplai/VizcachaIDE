// The three language profiles as the backend declares them (adapters/golang/profile.go and the
// provisional profiles of Python and C++, docs/PLAN_PYTHON.md and docs/PLAN_CPP.md section 4.1).
// The mock serves them; the Wails bridge uses them only until bridge.CodeLanguagesService exists
// (track N5), then this file serves the mock alone.
import type { LanguageProfile, ToolSpec } from '../domain'

const tool = (spec: Partial<ToolSpec> & Pick<ToolSpec, 'id' | 'role'>): ToolSpec => ({
  labelKey: '',
  missingKey: '',
  installUrl: '',
  installCommand: '',
  providedBy: '',
  ...spec
})

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
  tools: []
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
    check: false,
    debugInput: false,
    packageActions: [],
    threadsLabel: 'debug.threads'
  },
  tools: []
}

export const languageProfiles: LanguageProfile[] = [goProfile, pythonProfile, cppProfile]
