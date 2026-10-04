// Settings and detected tools of the mock bridge (see mock.ts).
import type { CodeLanguage, Settings, ToolSource, ToolStatus } from '../domain'
import { resolveLanguage, systemLanguage } from '../language'
import type { Emit } from './mockScenarios'
import type { Bridge } from './types'

/** The part of the mock state that holds the user's settings. */
export interface SettingsHolder {
  settings: Settings
}

const sourceOf = (settings: Settings, toolId: string, found: ToolSource): ToolSource =>
  settings.toolPaths[toolId] ? 'configured' : found

const detected = (
  settings: Settings,
  codeLanguage: CodeLanguage,
  id: string,
  role: ToolStatus['role'],
  version: string,
  found: ToolSource
): ToolStatus => ({
  id,
  codeLanguage,
  role,
  version,
  source: sourceOf(settings, id, found),
  path: settings.toolPaths[id] ?? ''
})

/** The detected tools of every language. */
export const toolsFor = (settings: Settings): ToolStatus[] => [
  detected(settings, 'go', 'go', 'runtime', '1.25.5', 'bundled'),
  detected(settings, 'go', 'dlv', 'debugAdapter', '1.27.2', 'bundled'),
  detected(settings, 'go', 'gopls', 'languageServer', '0.21.1', 'path'),
  detected(settings, 'python', 'python', 'runtime', '3.12.4', 'path'),
  // The modules live inside the interpreter: they have no path of their own.
  detected(settings, 'python', 'debugpy', 'debugAdapter', '1.8.5', 'path'),
  detected(settings, 'python', 'pylsp', 'languageServer', '1.12.0', 'path'),
  detected(settings, 'python', 'ruff', 'formatter', '0.6.9', 'path'),
  detected(settings, 'cpp', 'cxx', 'compiler', '15.2.0', 'path'),
  detected(settings, 'cpp', 'lldb-dap', 'debugAdapter', '23.1.2', 'path'),
  detected(settings, 'cpp', 'clangd', 'languageServer', '23.1.2', 'path'),
  detected(settings, 'cpp', 'clang-format', 'formatter', '23.1.2', 'path')
]

export const mockSettings = (state: SettingsHolder, emit: Emit): Bridge['settings'] => {
  const save = async (next: Settings): Promise<void> => {
    state.settings = next
    emit('settings:changed', next)
  }
  return {
    get: async () => state.settings,
    save,
    pickExecutable: async (tool) => {
      const toolPaths = { ...state.settings.toolPaths, [tool]: `C:\\tools\\${tool}.exe` }
      await save({ ...state.settings, toolPaths })
      return toolsFor(state.settings)
    },
    resolvedLanguage: async () => resolveLanguage(state.settings.language, systemLanguage())
  }
}
