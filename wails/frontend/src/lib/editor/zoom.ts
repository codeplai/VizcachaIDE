import { derived, get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { settings, updateSettings } from '../stores/settings'

export const DEFAULT_FONT_SIZE = 14
export const MIN_FONT_SIZE = 10
export const MAX_FONT_SIZE = 32

export type ZoomAction = 'in' | 'out' | 'reset'

export const nextFontSize = (current: number, action: ZoomAction): number => {
  if (action === 'reset') return DEFAULT_FONT_SIZE
  const step = action === 'in' ? 1 : -1
  return Math.min(MAX_FONT_SIZE, Math.max(MIN_FONT_SIZE, current + step))
}

/** Editor text size in pixels: the one saved in Settings, 14 until they load. */
export const editorFontSize = derived(settings, (value) => value?.fontSize || DEFAULT_FONT_SIZE)

/** Ctrl+= / Ctrl+- / Ctrl+0. The size is saved in Settings, so it survives restarts. */
export const zoomEditor = async (bridge: Bridge, action: ZoomAction): Promise<void> => {
  await updateSettings(bridge, { fontSize: nextFontSize(get(editorFontSize), action) })
}
