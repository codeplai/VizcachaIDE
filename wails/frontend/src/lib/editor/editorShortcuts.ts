// Window-level editor shortcuts: Ctrl+S saves (formatting first); Ctrl+= / Ctrl+- / Ctrl+0 zoom.
import type { Bridge } from '../bridge'
import { saveActiveDocument } from './tabs'
import { zoomEditor, type ZoomAction } from './zoom'

const ZOOM_KEYS: Record<string, ZoomAction> = {
  '=': 'in',
  '+': 'in',
  '-': 'out',
  _: 'out',
  '0': 'reset'
}

export const registerEditorShortcuts = (bridge: Bridge): (() => void) => {
  const onKey = (event: KeyboardEvent): void => {
    if (!(event.ctrlKey || event.metaKey) || event.altKey) return
    if (event.key.toLowerCase() === 's') {
      event.preventDefault()
      void saveActiveDocument(bridge)
      return
    }
    const zoom = ZOOM_KEYS[event.key]
    if (!zoom) return
    event.preventDefault()
    void zoomEditor(bridge, zoom)
  }
  const onWheel = (event: WheelEvent): void => {
    if (!(event.ctrlKey || event.metaKey)) return
    event.preventDefault()
    void zoomEditor(bridge, event.deltaY < 0 ? 'in' : 'out')
  }
  window.addEventListener('keydown', onKey)
  window.addEventListener('wheel', onWheel, { passive: false })
  return () => {
    window.removeEventListener('keydown', onKey)
    window.removeEventListener('wheel', onWheel)
  }
}
