// File > New project: the dialog's state and the flow that creates the project and opens it.
import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage } from '../domain'
import { startingCodeLanguage } from './enabledLanguages'
import { openFile, showFolder } from './files'
import { openDialog } from './layout'
import { askNativeDialog } from './nativeDialogs'
import { settings } from './settings'

export const newProjectName = writable('')
export const newProjectLanguage = writable<CodeLanguage>('go')
/** The folder the project goes in. It is always asked: there is no default. */
export const newProjectLocation = writable('')
/** The i18n key of what went wrong, with the backend's own text for an unexpected failure. */
export const newProjectError = writable<{ key: string; reason: string } | null>(null)
export const newProjectBusy = writable(false)

const FORBIDDEN_CHARACTERS = /[\\/:*?"<>|]/
const ERROR_KEY = /project\.error[A-Za-z]+/

/** The i18n key of the problem a project name has, or null. Mirrors the backend's checks. */
export const projectNameProblem = (name: string): string | null => {
  const trimmed = name.trim()
  if (trimmed === '') return 'project.errorNameEmpty'
  if (FORBIDDEN_CHARACTERS.test(trimmed) || /^[. ]+$/.test(trimmed) || trimmed.endsWith('.')) {
    return 'project.errorNameInvalid'
  }
  return null
}

/** The name is not shown as an error while it is empty: Create is just disabled. */
export const visibleNameProblem = derived(newProjectName, (name) => {
  const problem = projectNameProblem(name)
  return problem === 'project.errorNameEmpty' ? null : problem
})

/** Create is enabled with a valid name and a location, and while nothing is being created. */
export const canCreateProject = derived(
  [newProjectName, newProjectLocation, newProjectBusy],
  ([name, location, busy]) => !busy && location !== '' && projectNameProblem(name) === null
)

/** Opens the dialog with an empty form, in the language new files start in. */
export const startNewProject = async (_bridge?: Bridge): Promise<void> => {
  newProjectName.set('')
  newProjectLocation.set('')
  newProjectError.set(null)
  newProjectBusy.set(false)
  newProjectLanguage.set(startingCodeLanguage(get(settings)?.defaultCodeLanguage))
  openDialog.set('newProject')
}

export const closeNewProject = (): void => {
  if (get(openDialog) === 'newProject') openDialog.set(null)
}

/** The native folder picker; keeps the current location when the student cancels. */
export const chooseProjectLocation = async (bridge: Bridge): Promise<void> => {
  const chosen = await askNativeDialog(() => bridge.files.chooseFolder(), '')
  if (!chosen) return
  newProjectLocation.set(chosen)
  newProjectError.set(null)
}

const failureOf = (error: unknown): { key: string; reason: string } => {
  const message = error instanceof Error ? error.message : String(error ?? '')
  const key = ERROR_KEY.exec(message)?.[0]
  return key ? { key, reason: '' } : { key: 'project.errorFailed', reason: message }
}

/**
 * Creates the project, shows its folder in the file tree (as Open folder does) and opens its main
 * file, ready to run. A failure stays in the dialog, which keeps what the student typed.
 */
export const createNewProject = async (bridge: Bridge): Promise<void> => {
  if (!get(canCreateProject)) return
  newProjectBusy.set(true)
  newProjectError.set(null)
  try {
    const project = await bridge.projects.create(
      get(newProjectLanguage),
      get(newProjectLocation),
      get(newProjectName).trim()
    )
    await showFolder(bridge, await bridge.files.listTree(project.root))
    await openFile(bridge, project.mainFile)
    closeNewProject()
  } catch (error) {
    newProjectError.set(failureOf(error))
  } finally {
    newProjectBusy.set(false)
  }
}
