// Open-file tabs as the tab bar shows them: name, modified dot and the active one.
import { derived } from 'svelte/store'
import { activePath, baseName, dirty, openTabs } from '../stores/files'

export interface TabItem {
  path: string
  name: string
  modified: boolean
  active: boolean
}

export const tabItems = derived([openTabs, activePath, dirty], ([paths, active, changed]) =>
  paths.map((path): TabItem => ({
    path,
    name: baseName(path),
    modified: !!changed[path],
    active: path === active
  }))
)
