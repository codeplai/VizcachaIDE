import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { setupI18n } from '../i18n'
import {
  activePath,
  fileTree,
  isValidModuleName,
  isValidPackage,
  lastRunConfiguration,
  moduleTask,
  openDialog,
  resetRun
} from '../stores'
import ModulesDialog from './ModulesDialog.svelte'

beforeAll(() => {
  setupI18n('en')
})

beforeEach(() => {
  resetRun()
  moduleTask.set(null)
  lastRunConfiguration.set(null)
  fileTree.set(null)
  activePath.set('C:/work/My Project/main.go')
  openDialog.set('modules')
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
    expect(isValidModuleName('example.com/hola')).toBe(true)
    expect(isValidModuleName('hola go')).toBe(false)
    expect(isValidModuleName('-x')).toBe(false)
    expect(isValidModuleName('')).toBe(false)
  })

  it('accepts packages with an optional version', () => {
    expect(isValidPackage('github.com/user/pkg')).toBe(true)
    expect(isValidPackage('github.com/user/pkg@v1.2.3')).toBe(true)
    expect(isValidPackage('github.com/user/pkg extra')).toBe(false)
  })
})

describe('Go modules dialog actions', () => {
  it('creates go.mod with the folder name prefilled', async () => {
    const init = vi.spyOn(bridge.run, 'modInit')
    render(ModulesDialog)
    const input = (await screen.findByLabelText('Module name')) as HTMLInputElement
    expect(input.value).toBe('my-project')
    await fireEvent.click(screen.getByRole('button', { name: 'Create go.mod' }))
    expect(init).toHaveBeenCalledWith('C:/work/My Project', 'my-project')
  })

  it('does not create go.mod with an invalid name', async () => {
    render(ModulesDialog)
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
    const tidy = vi.spyOn(bridge.run, 'modTidy')
    const get = vi.spyOn(bridge.run, 'modGet')
    const open = vi.spyOn(bridge.system, 'openUrl').mockImplementation(() => {})
    render(ModulesDialog)
    await fireEvent.click(await screen.findByRole('button', { name: 'Tidy up dependencies' }))
    expect(tidy).toHaveBeenCalledWith('C:/work/My Project')

    await fireEvent.input(screen.getByLabelText('Package to add'), {
      target: { value: 'github.com/user/pkg' }
    })
    await fireEvent.click(screen.getByRole('button', { name: 'Add a package' }))
    expect(get).toHaveBeenCalledWith('C:/work/My Project', 'github.com/user/pkg')

    await fireEvent.click(screen.getByRole('button', { name: 'Search packages on pkg.go.dev' }))
    expect(open).toHaveBeenCalledWith('https://pkg.go.dev')
  })
})
