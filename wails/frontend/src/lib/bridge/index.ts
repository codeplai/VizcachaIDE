// The one door to the backend. Components and stores import from here, never from wailsjs.
import { createMockBridge, type MockControls } from './mock'
import type { Bridge } from './types'
import { createWailsBridge, hasWailsBackend } from './wails'

export type { Bridge, ToolId, Unsubscribe } from './types'
export type { MockControls } from './mock'
export type { Scenario } from './mockScenarios'

export interface BridgeSelection {
  bridge: Bridge
  /** Only present with the mock backend (browser development). */
  mockControls: MockControls | null
}

/** Picks the Wails backend when `window.go` exists, the mock otherwise. */
export const selectBridge = (): BridgeSelection => {
  if (hasWailsBackend()) return { bridge: createWailsBridge(), mockControls: null }
  const mock = createMockBridge()
  return { bridge: mock.bridge, mockControls: mock.controls }
}

const selection = selectBridge()

export const bridge: Bridge = selection.bridge
export const mockControls: MockControls | null = selection.mockControls
