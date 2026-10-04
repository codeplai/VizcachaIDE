import { cleanup, render, screen } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { goProfile, pythonProfile } from '../bridge/languageProfiles'
import type { LanguageProfile, ToolSpec } from '../domain'
import { setupI18n } from '../i18n'
import { profiles, tools } from '../stores'
import SettingsTools from './SettingsTools.svelte'

const spec = (id: string, providedBy: string): ToolSpec => ({
  id,
  role: providedBy ? 'debugAdapter' : 'runtime',
  labelKey: `test.${id}`,
  missingKey: '',
  installUrl: '',
  installCommand: '',
  providedBy
})

const python: LanguageProfile = {
  ...pythonProfile,
  tools: [spec('python', ''), spec('debugpy', 'python')]
}

beforeAll(() => setupI18n('en'))

beforeEach(() => {
  profiles.set([goProfile, python])
  tools.set([
    {
      id: 'python',
      codeLanguage: 'python',
      role: 'runtime',
      version: '3.13.1',
      source: 'path',
      path: ''
    },
    {
      id: 'debugpy',
      codeLanguage: 'python',
      role: 'debugAdapter',
      version: '1.8.9',
      source: 'path',
      path: ''
    }
  ])
})

afterEach(() => {
  cleanup()
  profiles.set([])
  tools.set([])
})

describe('Settings tools', () => {
  it('has one group per profile and one row per tool', () => {
    render(SettingsTools)
    expect(screen.getAllByRole('heading')).toHaveLength(2)
    expect(document.getElementById('setting-go')).not.toBeNull()
    expect(document.getElementById('setting-dlv')).not.toBeNull()
    expect(document.getElementById('setting-gopls')).not.toBeNull()
    expect(document.getElementById('setting-python')).not.toBeNull()
  })

  it('shows state and version of a tool provided by another, without a path or a Choose button', () => {
    render(SettingsTools)
    expect(screen.getByText('test.debugpy')).toBeTruthy()
    expect(screen.getByText(/1\.8\.9/)).toBeTruthy()
    expect(document.getElementById('setting-debugpy')).toBeNull()
    // Go has three tools with their own path, Python one: the provided one has no button.
    expect(screen.getAllByRole('button', { name: 'Choose…' })).toHaveLength(4)
  })
})
