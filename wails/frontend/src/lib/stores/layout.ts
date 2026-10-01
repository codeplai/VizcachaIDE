import { writable } from 'svelte/store'

export type SidebarView = 'files' | 'outline' | 'search'
export type OutputTab = 'output' | 'problems'

export const sidebarView = writable<SidebarView>('files')
export const outputTab = writable<OutputTab>('output')
/** Cursor position shown in the status bar (1-based). */
export const cursor = writable({ line: 1, column: 1 })
