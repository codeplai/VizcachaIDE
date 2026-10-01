// Settings and toolchain part of the mock bridge (see mock.ts).
import type { Settings, ToolchainInfo } from '../domain'
import { resolveLanguage, systemLanguage } from '../language'
import type { Emit } from './mockScenarios'
import type { Bridge } from './types'

/** The part of the mock state that holds the user's settings. */
export interface SettingsHolder {
  settings: Settings
}

const TOOL_FIELDS = { go: 'goPath', dlv: 'delvePath', gopls: 'goplsPath' } as const

export const toolchainFor = (settings: Settings): ToolchainInfo => ({
  goVersion: '1.25.5',
  delveVersion: '1.27.2',
  goplsVersion: '0.21.1',
  goSource: settings.goPath ? 'configured' : 'bundled',
  delveSource: settings.delvePath ? 'configured' : 'bundled',
  goplsSource: settings.goplsPath ? 'configured' : 'path'
})

export const mockSettings = (state: SettingsHolder, emit: Emit): Bridge['settings'] => {
  const save = async (next: Settings): Promise<void> => {
    state.settings = next
    emit('settings:changed', next)
  }
  return {
    get: async () => state.settings,
    save,
    pickExecutable: async (tool) => {
      await save({ ...state.settings, [TOOL_FIELDS[tool]]: `C:\\tools\\${tool}.exe` })
      return toolchainFor(state.settings)
    },
    resolvedLanguage: async () => resolveLanguage(state.settings.language, systemLanguage())
  }
}
