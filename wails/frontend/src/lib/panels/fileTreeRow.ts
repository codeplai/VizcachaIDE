// Mouse, keyboard and drag events of one row of the Files panel.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { FileNode } from '../domain'
import {
  canDrop,
  clearSelection,
  copyEntries,
  cutEntries,
  deleteSelected,
  dragPaths,
  dropEntries,
  dropTarget,
  openFile,
  pasteEntries,
  selectEntry,
  selectedPaths,
  startDrag,
  startRename,
  toggleFolder
} from '../stores'

export interface RowEvents {
  onClick: (event: MouseEvent) => void
  onKeydown: (event: KeyboardEvent) => void
  onDragStart: (event: DragEvent) => void
  onDragOver: (event: DragEvent) => void
  onDrop: (event: DragEvent) => void
}

/** Ctrl/Cmd+X, C and V: only on a row, so the panel has the focus (the editor and terminal keep theirs). */
const clipboardKey = (
  bridge: Bridge,
  node: FileNode,
  depth: number,
  event: KeyboardEvent
): boolean => {
  if (!(event.ctrlKey || event.metaKey) || event.shiftKey || event.altKey) return false
  const key = event.key.toLowerCase()
  if (key === 'v') void pasteEntries(bridge, node.isDir ? node.path : undefined)
  else if (key === 'x' && depth > 0) cutEntries(node.path)
  else if (key === 'c' && depth > 0) copyEntries(node.path)
  else return false
  event.preventDefault()
  return true
}

export const rowEvents = (bridge: Bridge, node: FileNode, depth: number): RowEvents => ({
  onClick: (event) => {
    const toggle = event.ctrlKey || event.metaKey
    selectEntry(node.path, { toggle, range: event.shiftKey })
    if (node.isDir && !toggle && !event.shiftKey) toggleFolder(node.path)
  },
  onKeydown: (event) => {
    if (clipboardKey(bridge, node, depth, event)) return
    if (event.key === 'F2' && depth > 0) {
      event.preventDefault()
      startRename(node.path)
    } else if (event.key === 'Delete' && depth > 0) {
      event.preventDefault()
      void deleteSelected(bridge, node.path)
    } else if (event.key === 'Escape') {
      clearSelection()
    } else if (event.key === 'Enter' && !node.isDir) {
      event.preventDefault()
      void openFile(bridge, node.path)
    }
  },
  onDragStart: (event) => {
    if (depth === 0 || !event.dataTransfer) {
      event.preventDefault()
      return
    }
    startDrag(node.path, get(selectedPaths))
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', node.name)
  },
  onDragOver: (event) => {
    if (!node.isDir || !canDrop(get(dragPaths), node.path)) return
    event.preventDefault()
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
    dropTarget.set(node.path)
  },
  onDrop: (event) => {
    if (!node.isDir) return
    event.preventDefault()
    void dropEntries(bridge, node.path)
  }
})
