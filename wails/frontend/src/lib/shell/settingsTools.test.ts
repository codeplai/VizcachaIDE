import { cleanup, render, screen } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { cppProfile, goProfile, pythonProfile } from '../bridge/languageProfiles'
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
  profiles.set([goProfile, python, cppProfile])
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
  it('has one group per profile with tools and one row per tool', () => {
    render(SettingsTools)
    // Go, Python and C++ each have a group.
    expect(screen.getAllByRole('heading')).toHaveLength(3)
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
    // Go has three tools with their own path, Python one, C++ seven: the provided one has none.
    expect(screen.getAllByRole('button', { name: 'Choose…' })).toHaveLength(11)
  })
})
