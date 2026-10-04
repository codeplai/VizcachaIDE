// Updates part of the mock bridge (see mock.ts). `?update=ready` in the dev bar starts with a
// downloaded version, so the notice and the Settings tab can be seen without a release.
import type { UpdateState } from '../domain'
import type { Emit } from './mockScenarios'
import type { UpdatesApi } from './types'

const demoRelease = {
  version: '9.9.9',
  notes: 'Demo release notes.',
  notesUrl: 'https://github.com/codeplai/VizcachaIDE/releases',
  publishedAt: '2026-10-04T12:00:00Z',
  asset: 'VizcachaIDE-9.9.9-windows-amd64-full-setup.exe',
  assetSize: 1000
}

const initialState = (): UpdateState => {
  const ready =
    typeof location !== 'undefined' &&
    new URLSearchParams(location.search).get('update') === 'ready'
  return {
    status: ready ? 'ready' : 'idle',
    current: __APP_VERSION__,
    latest: ready ? demoRelease : null,
    downloadedBytes: ready ? 1000 : 0,
    totalBytes: ready ? 1000 : 0,
    installs: true,
    checkedAt: '',
    error: ''
  }
}

export const mockUpdates = (emit: Emit): UpdatesApi => {
  let state = initialState()
  const set = (change: Partial<UpdateState>): UpdateState => {
    state = { ...state, ...change }
    emit('update:state', state)
    return state
  }
  return {
    state: async () => state,
    check: async () =>
      state.status === 'ready'
        ? state
        : set({ status: 'upToDate', checkedAt: new Date().toISOString() }),
    download: async () => {},
    install: async () => {}
  }
}
