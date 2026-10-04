// The update of the IDE: what the backend reports (update:state) and the notice when a new version
// is ready. Checking at start, once a day, is the backend's job (app.AutoUpdate).
import { writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { UpdateState } from '../domain'
import { showNotice } from './notice'

export const updateState = writable<UpdateState | null>(null)

/** Versions already announced, so the notice appears once per version. */
const announced = new Set<string>()

/** "Check now": asks for the latest release and downloads it when there is one. */
export const checkForUpdates = async (bridge: Bridge): Promise<void> => {
  const state = await bridge.updates.check()
  updateState.set(state)
  if (state.status === 'available') await bridge.updates.download()
}

export const installUpdate = (bridge: Bridge): Promise<void> => bridge.updates.install()

const announceReady = (bridge: Bridge, state: UpdateState): void => {
  const latest = state.latest
  if (state.status !== 'ready' || !latest || announced.has(latest.version)) return
  announced.add(latest.version)
  showNotice({
    tone: 'info',
    messageKey: state.installs ? 'updates.readyInstall' : 'updates.readyFile',
    values: { version: latest.version },
    actions: [
      {
        labelKey: state.installs ? 'updates.installRestart' : 'updates.showFile',
        run: () => void installUpdate(bridge)
      },
      { labelKey: 'updates.whatsNew', run: () => bridge.system.openUrl(latest.notesUrl) }
    ]
  })
}

/** Follows update:state and announces a downloaded version. */
export const connectUpdates = async (bridge: Bridge): Promise<Unsubscribe> => {
  const off = bridge.on('update:state', (state) => {
    updateState.set(state)
    announceReady(bridge, state)
  })
  const initial = await bridge.updates.state()
  updateState.set(initial)
  announceReady(bridge, initial)
  return off
}
