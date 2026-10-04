import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'
import { bridge as appBridge } from '../bridge'
import { createMockBridge } from '../bridge/mock'
import type { UpdateState } from '../domain'
import { setupI18n } from '../i18n'
import SettingsUpdates from '../shell/SettingsUpdates.svelte'
import { notice, dismissNotice } from './notice'
import { connectSettings, settings } from './settings'
import { checkForUpdates, connectUpdates, updateState } from './updates'

const release = {
  version: '2.3.0',
  notes: 'Notes',
  notesUrl: 'https://example.org/2.3.0',
  publishedAt: '2026-10-04T12:00:00Z',
  asset: 'VizcachaIDE-2.3.0-windows-amd64-full-setup.exe',
  assetSize: 100
}

const stateOf = (change: Partial<UpdateState>): UpdateState => ({
  status: 'idle',
  current: '2.2.0',
  latest: null,
  downloadedBytes: 0,
  totalBytes: 0,
  installs: true,
  checkedAt: '',
  error: '',
  ...change
})

beforeAll(() => setupI18n('en'))

afterEach(() => {
  cleanup()
  dismissNotice()
  updateState.set(null)
  vi.restoreAllMocks()
})

describe('updates', () => {
  it('announces a downloaded version once, with Install and restart', async () => {
    const { bridge } = createMockBridge()
    const listeners: ((state: UpdateState) => void)[] = []
    vi.spyOn(bridge, 'on').mockImplementation(((
      _name: string,
      handler: (state: UpdateState) => void
    ) => {
      listeners.push(handler)
      return () => undefined
    }) as typeof bridge.on)
    const off = await connectUpdates(bridge)
    const ready = stateOf({
      status: 'ready',
      latest: release,
      downloadedBytes: 100,
      totalBytes: 100
    })
    listeners.forEach((listener) => listener(ready))
    expect(get(notice)?.messageKey).toBe('updates.readyInstall')
    expect(get(notice)?.actions.map((action) => action.labelKey)).toEqual([
      'updates.installRestart',
      'updates.whatsNew'
    ])
    dismissNotice()
    listeners.forEach((listener) => listener(ready))
    expect(get(notice)).toBeNull()
    off()
  })

  it('a portable copy is offered the downloaded file instead of installing', async () => {
    const { bridge } = createMockBridge()
    vi.spyOn(bridge.updates, 'state').mockResolvedValue(
      stateOf({ status: 'ready', latest: { ...release, version: '2.4.0' }, installs: false })
    )
    const off = await connectUpdates(bridge)
    expect(get(notice)?.messageKey).toBe('updates.readyFile')
    expect(get(notice)?.actions[0]?.labelKey).toBe('updates.showFile')
    off()
  })

  it('"Check now" downloads a version it finds', async () => {
    const { bridge } = createMockBridge()
    vi.spyOn(bridge.updates, 'check').mockResolvedValue(
      stateOf({ status: 'available', latest: release })
    )
    const download = vi.spyOn(bridge.updates, 'download').mockResolvedValue()
    await checkForUpdates(bridge)
    expect(download).toHaveBeenCalledTimes(1)
    expect(get(updateState)?.status).toBe('available')
  })

  it('the Settings tab shows the progress, the install button and the automatic check', async () => {
    // The component saves through the app's bridge, so the settings store follows that one.
    const offSettings = await connectSettings(appBridge)
    updateState.set(
      stateOf({ status: 'downloading', latest: release, downloadedBytes: 42, totalBytes: 100 })
    )
    render(SettingsUpdates)
    expect(screen.getByText('Downloading version 2.3.0… 42 %')).toBeTruthy()
    expect(document.querySelector('progress')?.getAttribute('value')).toBe('42')
    updateState.set(stateOf({ status: 'ready', latest: release }))
    expect(await screen.findByRole('button', { name: 'Install and restart' })).toBeTruthy()
    await fireEvent.click(screen.getByRole('checkbox'))
    await vi.waitFor(() => expect(get(settings)?.checkUpdates).toBe(false))
    offSettings()
  })
})
