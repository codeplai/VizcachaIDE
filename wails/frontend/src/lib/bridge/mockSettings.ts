// Settings and toolchain part of the mock bridge (see mock.ts).
import type { Settings, ToolchainInfo, ToolSource, ToolStatus } from '../domain'
import { resolveLanguage, systemLanguage } from '../language'
import type { Emit } from './mockScenarios'
import type { Bridge } from './types'

/** The part of the mock state that holds the user's settings. */
export interface SettingsHolder {
  settings: Settings
}

const sourceOf = (settings: Settings, toolId: string, found: ToolSource): ToolSource =>
  settings.toolPaths[toolId] ? 'configured' : found

/** Transitional (M0): the old shape of the detected tools; see toolsFor. */
export const toolchainFor = (settings: Settings): ToolchainInfo => ({
  goVersion: '1.25.5',
  delveVersion: '1.27.2',
  goplsVersion: '0.21.1',
  goSource: sourceOf(settings, 'go', 'bundled'),
  delveSource: sourceOf(settings, 'dlv', 'bundled'),
  goplsSource: sourceOf(settings, 'gopls', 'path')
})

const goTool = (
  settings: Settings,
  id: string,
  role: ToolStatus['role'],
  version: string,
  found: ToolSource
): ToolStatus => ({
  id,
  codeLanguage: 'go',
  role,
  version,
  source: sourceOf(settings, id, found),
  path: settings.toolPaths[id] ?? ''
})

/** The detected tools of every language (only Go has tools in the demo). */
export const toolsFor = (settings: Settings): ToolStatus[] => [
  goTool(settings, 'go', 'runtime', '1.25.5', 'bundled'),
  goTool(settings, 'dlv', 'debugAdapter', '1.27.2', 'bundled'),
  goTool(settings, 'gopls', 'languageServer', '0.21.1', 'path')
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
      return toolchainFor(state.settings)
    },
    resolvedLanguage: async () => resolveLanguage(state.settings.language, systemLanguage())
  }
}
