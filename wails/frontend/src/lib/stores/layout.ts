import { writable } from 'svelte/store'

export type SidebarView = 'files' | 'outline' | 'search'
export type OutputTab = 'output' | 'problems'
export type DebugTab = 'stack' | 'goroutines'
export type DialogName = 'settings' | 'about' | 'modules'
export type SettingsTab = 'general' | 'editor' | 'tools'

export const sidebarView = writable<SidebarView>('files')
export const outputTab = writable<OutputTab>('output')
export const debugTab = writable<DebugTab>('stack')
export const settingsTab = writable<SettingsTab>('general')
/** The dialog that is open, if any (confirmations have their own store). */
export const openDialog = writable<DialogName | null>(null)
/** Folders the user closed in the file tree. Everything else is shown open. */
export const collapsedFolders = writable<Set<string>>(new Set())
/** The tree row that is highlighted (a click selects, a double click opens). */
export const selectedNode = writable<string | null>(null)
/** Cursor position shown in the status bar (1-based). */
export const cursor = writable({ line: 1, column: 1 })

export const toggleFolder = (path: string): void =>
  collapsedFolders.update((all) => {
    const next = new Set(all)
    if (!next.delete(path)) next.add(path)
    return next
  })
