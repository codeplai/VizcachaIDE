import { get, writable } from 'svelte/store'
import { capabilities } from './codeLanguages'

export type SidebarView = 'files' | 'outline' | 'search'
export type OutputTab = 'output' | 'problems' | 'console' | 'terminal'
export type DebugTab = 'stack' | 'calls' | 'threads'
export type DialogName = 'settings' | 'about' | 'packages' | 'newProject'
export type SettingsTab = 'general' | 'editor' | 'tools' | 'updates'

export const sidebarView = writable<SidebarView>('files')
export const outputTab = writable<OutputTab>('output')
/** The tab of the debug side panel. */
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

// The Console tab exists only for languages with a console: leave it when the file changes to one without.
capabilities.subscribe((current) => {
  if (current && !current.console && get(outputTab) === 'console') outputTab.set('output')
})

export const toggleFolder = (path: string): void =>
  collapsedFolders.update((all) => {
    const next = new Set(all)
    if (!next.delete(path)) next.add(path)
    return next
  })

/**
 * Whether the side panel (Files, Outline, Search) and the Assistant are shown. Hiding them gives
 * the editor more room; the panes remember it between sessions (paneforge saves collapsed panes).
 */
export const sidebarOpen = writable(true)
export const assistantOpen = writable(true)

export const toggleSidebar = (): void => sidebarOpen.update((open) => !open)
export const toggleAssistant = (): void => assistantOpen.update((open) => !open)

/** A click on a rail icon: the active one hides or shows the panel, another one switches to it. */
export const selectSidebarView = (view: SidebarView): void => {
  if (get(sidebarView) === view && get(sidebarOpen)) {
    sidebarOpen.set(false)
    return
  }
  sidebarView.set(view)
  sidebarOpen.set(true)
}
