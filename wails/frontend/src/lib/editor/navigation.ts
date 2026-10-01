// "Go to this line" requests from other panels (Problems, Outline, Output links, definitions).
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { SourceLocation } from '../domain'
import { activePath, openFile } from '../stores/files'

export interface GotoRequest {
  file: string
  line: number
  column: number
  /** Changes on every request, so asking for the same line twice still moves the cursor. */
  nonce: number
}

export const gotoRequest = writable<GotoRequest | null>(null)

let counter = 0

/** Opens (or activates) the file's tab and puts the cursor on the location. */
export const goToLocation = async (bridge: Bridge, location: SourceLocation): Promise<void> => {
  if (get(activePath) !== location.file) await openFile(bridge, location.file)
  gotoRequest.set({ ...location, nonce: ++counter })
}
