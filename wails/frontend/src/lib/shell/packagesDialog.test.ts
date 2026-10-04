import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles } from '../bridge/languageProfiles'
import { setupI18n } from '../i18n'
import {
  activePath,
  fileTree,
  isValidProjectName,
  isValidPackage,
  lastRunConfiguration,
  openDialog,
  packageTask,
  profiles,
  resetRun
} from '../stores'
import PackagesDialog from './PackagesDialog.svelte'

beforeAll(() => {
  setupI18n('en')
})

beforeEach(() => {
  resetRun()
  profiles.set(languageProfiles)
  packageTask.set(null)
  lastRunConfiguration.set(null)
  fileTree.set(null)
  activePath.set('C:/work/My Project/main.go')
  openDialog.set('packages')
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  openDialog.set(null)
  fileTree.set(null)
  activePath.set(null)
})

describe('module name and package validation', () => {
  it('accepts module paths and rejects spaces and flags', () => {
    expect(isValidProjectName('example.com/hola')).toBe(true)
    expect(isValidProjectName('hola go')).toBe(false)
    expect(isValidProjectName('-x')).toBe(false)
    expect(isValidProjectName('')).toBe(false)
  })

  it('accepts packages with an optional version', () => {
    expect(isValidPackage('github.com/user/pkg')).toBe(true)
    expect(isValidPackage('github.com/user/pkg@v1.2.3')).toBe(true)
    expect(isValidPackage('github.com/user/pkg extra')).toBe(false)
  })
})

describe('Packages dialog actions for Go', () => {
  it('creates go.mod with the folder name prefilled', async () => {
    const init = vi.spyOn(bridge.packages, 'init')
    render(PackagesDialog)
    const input = (await screen.findByLabelText('Module name')) as HTMLInputElement
    expect(input.value).toBe('my-project')
    await fireEvent.click(screen.getByRole('button', { name: 'Create go.mod' }))
    expect(init).toHaveBeenCalledWith('go', 'C:/work/My Project', 'my-project')
  })

  it('does not create go.mod with an invalid name', async () => {
    render(PackagesDialog)
    const input = await screen.findByLabelText('Module name')
    await fireEvent.input(input, { target: { value: 'two words' } })
    const button = screen.getByRole('button', { name: 'Create go.mod' }) as HTMLButtonElement
    expect(button.disabled).toBe(true)
  })

  it('tidies and adds a package when the folder has a go.mod', async () => {
    fileTree.set({
      name: 'My Project',
      path: 'C:/work/My Project',
      isDir: true,
      children: [{ name: 'go.mod', path: 'C:/work/My Project/go.mod', isDir: false, children: [] }]
    })
    const tidy = vi.spyOn(bridge.packages, 'tidy')
    const get = vi.spyOn(bridge.packages, 'add')
    const open = vi.spyOn(bridge.system, 'openUrl').mockImplementation(() => {})
    render(PackagesDialog)
    await fireEvent.click(await screen.findByRole('button', { name: 'Tidy up dependencies' }))
    expect(tidy).toHaveBeenCalledWith('go', 'C:/work/My Project')

    await fireEvent.input(screen.getByLabelText('Package to add'), {
      target: { value: 'github.com/user/pkg' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Add a package' }))
    expect(get).toHaveBeenCalledWith('go', 'C:/work/My Project', 'github.com/user/pkg')

    await fireEvent.click(screen.getByRole('button', { name: 'Search packages on pkg.go.dev' }))
    expect(open).toHaveBeenCalledWith('https://pkg.go.dev')
  })
})

describe('Packages dialog for Python', () => {
  beforeEach(() => activePath.set('C:/work/app/main.py'))

  it('offers adding packages and no go.mod', async () => {
    const add = vi.spyOn(bridge.packages, 'add')
    render(PackagesDialog)
    expect(screen.queryByLabelText('Module name')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Tidy up dependencies' })).toBeNull()
    await fireEvent.input(await screen.findByLabelText('Package to add'), {
      target: { value: 'requests==2.32.0' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Install a package' }))
    expect(add).toHaveBeenCalledWith('python', 'C:/work/app', 'requests==2.32.0')
  })

  it('accepts pip requirement strings', () => {
    expect(isValidPackage('requests', 'python')).toBe(true)
    expect(isValidPackage('requests>=2.0', 'python')).toBe(true)
    expect(isValidPackage('two words', 'python')).toBe(false)
  })
})
