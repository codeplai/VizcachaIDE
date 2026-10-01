import { derived, get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { DEFAULT_FONT_SIZE, clampFontSize, settings, updateSettings } from '../stores/settings'

export type ZoomAction = 'in' | 'out' | 'reset'

export const nextFontSize = (current: number, action: ZoomAction): number => {
  if (action === 'reset') return DEFAULT_FONT_SIZE
  return clampFontSize(current + (action === 'in' ? 1 : -1))
}

/** Editor text size in pixels: the one saved in Settings, 14 until they load. */
export const editorFontSize = derived(settings, (value) => value?.fontSize || DEFAULT_FONT_SIZE)

/** Ctrl+= / Ctrl+- / Ctrl+0. The size is saved in Settings, so it survives restarts. */
export const zoomEditor = async (bridge: Bridge, action: ZoomAction): Promise<void> => {
  await updateSettings(bridge, { fontSize: nextFontSize(get(editorFontSize), action) })
}
