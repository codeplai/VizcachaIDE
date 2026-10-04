import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { bridge, mockControls, type Unsubscribe } from '../bridge'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  confirmReplaceAll,
  connectStores,
  debugActive,
  lastRunConfiguration,
  openDialog,
  openTabs,
  resetRun,
  settingsTab,
  settings
} from '../stores'
import AboutDialog from './AboutDialog.svelte'
import ConfirmHost from './ConfirmHost.svelte'
import FirstRunWizard from './FirstRunWizard.svelte'
import PackagesDialog from './PackagesDialog.svelte'
import MoreMenu from './MoreMenu.svelte'
import SettingsDialog from './SettingsDialog.svelte'
import StatusBar from './StatusBar.svelte'

let stop: Unsubscribe | undefined

beforeAll(() => {
  setupI18n('en')
})

beforeEach(async () => {
  openDialog.set(null)
  settingsTab.set('general')
  buffers.set({})
  openTabs.set([])
  activePath.set(null)
  stop = await connectStores(bridge)
  await mockControls?.play('write')
  resetRun()
  await bridge.settings.save({ ...(await bridge.settings.get()), language: 'en', firstRun: false })
})

afterEach(() => {
  cleanup()
  stop?.()
})

describe('More menu', () => {
  it('opens the settings from the menu', async () => {
    render(MoreMenu)
    const trigger = screen.getByRole('button', { name: /More/ })
    await fireEvent.keyDown(trigger, { key: 'Enter' })
    await fireEvent.click(await screen.findByRole('menuitem', { name: 'Settings' }))
    expect(get(openDialog)).toBe('settings')
  })
})

describe('Settings dialog', () => {
  it('changes the language, the font size and "format when saving"', async () => {
    openDialog.set('settings')
    render(SettingsDialog)
    const language = (await screen.findByLabelText('Language')) as HTMLSelectElement
    await fireEvent.change(language, { target: { value: 'es' } })
    await waitFor(() => expect(get(settings)?.language).toBe('es'))

    await fireEvent.click(screen.getByRole('tab', { name: 'Editor' }))
    const size = screen.getByLabelText('Font size') as HTMLInputElement
    await fireEvent.change(size, { target: { value: '99' } })
    await waitFor(() => expect(get(settings)?.fontSize).toBe(28))
    const format = screen.getByLabelText('Format when saving') as HTMLInputElement
    await fireEvent.click(format)
    await waitFor(() => expect(get(settings)?.formatOnSave).toBe(false))
  })

  it('saves the tool paths and says where each tool comes from', async () => {
    openDialog.set('settings')
    settingsTab.set('tools')
    render(SettingsDialog)
    expect(await screen.findAllByText(/Included with VizcachaIDE/)).toHaveLength(2)
    // gopls, Python and its three modules, and the four tools of C++.
    expect(screen.getAllByText(/Found on your PATH/)).toHaveLength(9)
    const go = screen.getByLabelText('Go')
    await fireEvent.change(go, { target: { value: ' C:\\go\\bin\\go.exe ' } })
    await waitFor(() => expect(get(settings)?.toolPaths['go']).toBe('C:\\go\\bin\\go.exe'))
    expect(await screen.findByText(/Location you chose/)).toBeTruthy()
    await bridge.settings.save({ ...(await bridge.settings.get()), toolPaths: {} })
  })
})

describe('About and Packages dialogs', () => {
  it('lists the tool versions', async () => {
    openDialog.set('about')
    render(AboutDialog)
    expect(await screen.findByText(/1.25.5/)).toBeTruthy()
    expect(screen.getAllByText('Included with VizcachaIDE')).toHaveLength(2)
    expect(screen.getByText('Marks Calderon')).toBeTruthy()
    expect(screen.getByText('CEO Codeplai')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'hola@codeplai.pe' })).toBeTruthy()
    expect(screen.getByText('Version 2.2.0')).toBeTruthy()
    expect(screen.getByText(/ABSOLUTELY NO WARRANTY/)).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Read the MIT license' })).toBeTruthy()
  })

  it('explains that there is no go.mod, then shows the module of the last run', async () => {
    lastRunConfiguration.set(null)
    openDialog.set('packages')
    render(PackagesDialog)
    expect(await screen.findByText(/No go\.mod found/)).toBeTruthy()
    lastRunConfiguration.set({
      codeLanguage: 'go',
      target: 'main.go',
      workingDir: '.',
      mode: 'project',
      programArgs: [],
      project: { root: 'C:\\proj', kind: 'gomod', name: 'example.com/hola' },
      echo: false
    })
    expect(await screen.findByText('example.com/hola')).toBeTruthy()
  })
})

describe('Confirmations', () => {
  it('shows the exact texts and returns the choice', async () => {
    render(ConfirmHost)
    const answer = confirmReplaceAll(12, 'main.go')
    expect(await screen.findByText('Replace 12 matches in main.go?')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Replace 12' }))
    expect(await answer).toBe('replace')
  })
})

describe('First start wizard', () => {
  it('walks through four steps and opens the example', async () => {
    await bridge.settings.save({ ...(await bridge.settings.get()), firstRun: true })
    render(FirstRunWizard)
    expect(await screen.findByText(/Welcome to VizcachaIDE/)).toBeTruthy()
    expect(screen.getByText('Step 1 of 4')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Español' }))
    await waitFor(() => expect(get(settings)?.language).toBe('es'))
    await fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(screen.getByText('Which programming languages will you use?')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(await screen.findByText('Checking Go… Go 1.25.5 is ready.')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Next' }))
    expect(screen.getByText('Open your first program')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Open hello example' }))
    await waitFor(() => expect(get(settings)?.firstRun).toBe(false))
    expect(get(activePath)).toBe('untitled/main.go')
  })
})

describe('Status bar', () => {
  it('shows gopls, the Go version, the position and the language, and turns sand while debugging', async () => {
    const view = render(StatusBar)
    expect(await screen.findByText('Code helper ready')).toBeTruthy()
    expect(screen.getByText('Go 1.25.5')).toBeTruthy()
    expect(screen.getByText('Line 1, column 1')).toBeTruthy()
    expect(view.container.querySelector('footer')?.classList.contains('debug')).toBe(false)
    await mockControls?.play('debug')
    await waitFor(() => expect(get(debugActive)).toBe(true))
    expect(await screen.findByText('● Debugging')).toBeTruthy()
    expect(view.container.querySelector('footer')?.classList.contains('debug')).toBe(true)
    await mockControls?.play('write')
  })
})
