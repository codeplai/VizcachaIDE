import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles } from '../bridge/languageProfiles'
import { setupI18n } from '../i18n'
import {
  activePath,
  connectSettings,
  fileTree,
  lastRunConfiguration,
  openDialog,
  packageTask,
  profiles,
  refreshTools,
  resetRun,
  settings,
  settingsTab
} from '../stores'
import FirstRunWizard from './FirstRunWizard.svelte'
import PackagesDialog from './PackagesDialog.svelte'
import SettingsDialog from './SettingsDialog.svelte'

beforeAll(() => setupI18n('en'))

let stopSettings: (() => void) | undefined

beforeEach(async () => {
  stopSettings = await connectSettings(bridge)
  profiles.set(languageProfiles)
  await refreshTools(bridge)
  // The mock keeps what was saved: start every test from the defaults.
  await bridge.settings.save({
    ...(await bridge.settings.get()),
    enabledCodeLanguages: [],
    defaultCodeLanguage: 'go',
    firstRun: false
  })
})

afterEach(() => {
  stopSettings?.()
  cleanup()
  vi.restoreAllMocks()
  openDialog.set(null)
  activePath.set(null)
  fileTree.set(null)
})

describe('Packages dialog verbs for Python', () => {
  beforeEach(() => {
    resetRun()
    packageTask.set(null)
    lastRunConfiguration.set(null)
    fileTree.set(null)
    activePath.set('C:/work/app/main.py')
    openDialog.set('packages')
  })

  it('shows the pip hint, add, uninstall and list from the capabilities', async () => {
    const remove = vi.spyOn(bridge.packages, 'remove')
    const list = vi.spyOn(bridge.packages, 'list')
    render(PackagesDialog)
    expect(await screen.findByText(/Packages are installed with pip/)).toBeTruthy()
    expect(screen.getByText('Packages')).toBeTruthy()
    await fireEvent.input(screen.getByLabelText('Package to add'), { target: { value: 'numpy' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Uninstall' }))
    expect(remove).toHaveBeenCalledWith('python', 'C:/work/app', 'numpy')
    await fireEvent.click(screen.getByRole('button', { name: 'Show installed packages' }))
    expect(list).toHaveBeenCalledWith('python', 'C:/work/app')
    expect(screen.queryByRole('button', { name: 'Tidy up dependencies' })).toBeNull()
  })

  it('keeps the Go texts exactly for Go', async () => {
    activePath.set('C:/work/app/main.go')
    fileTree.set({
      name: 'app',
      path: 'C:/work/app',
      isDir: true,
      children: [{ name: 'go.mod', path: 'C:/work/app/go.mod', isDir: false, children: [] }]
    })
    render(PackagesDialog)
    expect(await screen.findByRole('button', { name: 'Add a package' })).toBeTruthy()
    expect(screen.queryByText(/pip/)).toBeNull()
    expect(screen.queryByRole('button', { name: 'Uninstall' })).toBeNull()
  })
})

describe('Settings > Tools for Python', () => {
  it('lists python with Choose… and the modules as rows with status and version', async () => {
    openDialog.set('settings')
    settingsTab.set('tools')
    render(SettingsDialog)
    expect(await screen.findByText('Python tools')).toBeTruthy()
    expect(screen.getByLabelText('Python')).toBeTruthy()
    for (const module of ['debugpy', 'pylsp', 'ruff']) {
      expect(screen.getByText(`${module} module`)).toBeTruthy()
      expect(screen.queryByLabelText(`${module} module`)).toBeNull()
    }
    // One per executable: go, dlv, gopls, python, the seven of C++ and the four of Rust. The modules have none.
    expect(screen.getAllByRole('button', { name: 'Choose…' })).toHaveLength(15)
    expect(screen.getByText(/Version 3\.12\.4/)).toBeTruthy()
  })

  it('shows only the chosen languages', async () => {
    settings.set({ ...(await bridge.settings.get()), enabledCodeLanguages: ['go'] })
    openDialog.set('settings')
    settingsTab.set('tools')
    render(SettingsDialog)
    expect(await screen.findByText('Go tools')).toBeTruthy()
    expect(screen.queryByText('Python tools')).toBeNull()
  })
})

describe('Settings > General languages you use', () => {
  const boxes = (): HTMLInputElement[] =>
    screen
      .getAllByRole('checkbox')
      .filter((box): box is HTMLInputElement => box instanceof HTMLInputElement)

  it('has one checkbox per profile, all ticked while nothing was chosen', async () => {
    openDialog.set('settings')
    settingsTab.set('general')
    render(SettingsDialog)
    expect(await screen.findByRole('group', { name: 'Languages you use' })).toBeTruthy()
    expect(boxes().map((box) => box.checked)).toEqual([true, true, true, true])
  })

  it('saves the choice and never lets the last one go', async () => {
    openDialog.set('settings')
    settingsTab.set('general')
    render(SettingsDialog)
    await screen.findByRole('group', { name: 'Languages you use' })
    await fireEvent.click(boxes()[1] as HTMLInputElement)
    await fireEvent.click(boxes()[2] as HTMLInputElement)
    await fireEvent.click(boxes()[3] as HTMLInputElement)
    await waitFor(() => expect(get(settings)?.enabledCodeLanguages).toEqual(['go']))
    await waitFor(() => expect(boxes()[0]?.disabled).toBe(true))
  })
})

describe('First start wizard languages step', () => {
  const next = () => fireEvent.click(screen.getByRole('button', { name: 'Next' }))

  it('writes enabledCodeLanguages and checks the tools of the chosen languages', async () => {
    settings.set({ ...(await bridge.settings.get()), firstRun: true })
    render(FirstRunWizard)
    await next()
    expect(screen.getByText('Which programming languages will you use?')).toBeTruthy()
    const go = screen.getByRole('checkbox', { name: 'Go' })
    await fireEvent.click(go)
    await waitFor(() =>
      expect(get(settings)?.enabledCodeLanguages).toEqual(['python', 'cpp', 'rust'])
    )
    await next()
    expect(await screen.findByText(/Python · Version 3\.12\.4/)).toBeTruthy()
    expect(screen.queryByText(/Checking Go/)).toBeNull()
  })

  it('opens the first program in the first chosen language', async () => {
    settings.set({
      ...(await bridge.settings.get()),
      firstRun: true,
      enabledCodeLanguages: ['python'],
      defaultCodeLanguage: 'python'
    })
    render(FirstRunWizard)
    await next()
    await next()
    await next()
    await fireEvent.click(screen.getByRole('button', { name: 'Open hello example' }))
    await waitFor(() => expect(get(activePath)).toBe('untitled/main.py'))
  })
})
