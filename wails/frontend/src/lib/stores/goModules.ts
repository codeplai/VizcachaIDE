import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { ProjectContext } from '../domain'
import { activePath, baseName, fileTree } from './files'
import { lastRunConfiguration, runLines, runResult, running } from './run'

/** The module the dialog knows about. A "go mod" command's own run must not make it forget. */
const knownModule = writable<ProjectContext | null>(null)
/** Set when the dialog ran "go mod init" successfully for this folder. */
const createdIn = writable<string | null>(null)

export type ModuleAction = 'init' | 'tidy' | 'get'

/** The last command the dialog started and the error that kept it from starting, if any. */
export const moduleTask = writable<{ action: ModuleAction; error: string | null } | null>(null)

lastRunConfiguration.subscribe((config) => {
  if (config?.project?.kind === 'gomod') knownModule.set(config.project)
})

const parentOf = (path: string): string => path.replace(/[\\/][^\\/]*$/, '')

/** The folder the module commands run in: the open folder, else the active file's folder. */
export const modulesFolder = derived(
  [fileTree, activePath, lastRunConfiguration],
  ([tree, active, last]) => {
    if (tree?.path) return tree.path
    if (active) return parentOf(active)
    return last?.workingDir ?? ''
  }
)

/** True when go.mod exists in that folder (as far as the IDE has seen). */
export const hasGoMod = derived(
  [modulesFolder, fileTree, knownModule, createdIn],
  ([folder, tree, known, created]) => {
    if (!folder) return false
    if (created === folder) return true
    if (known && known.root === folder) return true
    return tree?.path === folder && tree.children.some((node) => node.name === 'go.mod')
  }
)

/** Suggested module name: the folder's name, made safe for a module path. */
export const suggestedModuleName = derived(modulesFolder, (folder) =>
  baseName(folder)
    .toLowerCase()
    .replace(/[^a-z0-9._/-]+/g, '-')
    .replace(/^[-.]+|-+$/g, '')
)

const MODULE_PATH = /^[A-Za-z0-9][A-Za-z0-9._~/-]*$/
const PACKAGE_PATH = /^[A-Za-z0-9][A-Za-z0-9._~/-]*(@[A-Za-z0-9._+-]+)?$/

/** True when `go mod init` can take the text as a module path. */
export const isValidModuleName = (name: string): boolean => MODULE_PATH.test(name.trim())

/** True when `go get` can take the text as a package (optionally with @version). */
export const isValidPackage = (name: string): boolean => PACKAGE_PATH.test(name.trim())

/** Output of the command: the lines the run store collected, without the "start" marker. */
export const moduleOutput = derived(runLines, (lines) =>
  lines.filter((line) => line.kind !== 'start').map((line) => line.text)
)

/** True while a command started from the dialog is running. */
export const moduleBusy = derived(
  [running, moduleTask],
  ([isRunning, task]) => isRunning && task !== null
)

// A successful "go mod init" means the folder now has a go.mod.
runResult.subscribe((result) => {
  if (get(moduleTask)?.action === 'init' && result?.exitCode === 0) {
    createdIn.set(get(modulesFolder))
  }
})

const start = (bridge: Bridge, action: ModuleAction, folder: string, argument: string) => {
  if (action === 'init') return bridge.run.modInit(folder, argument)
  if (action === 'get') return bridge.run.modGet(folder, argument)
  return bridge.run.modTidy(folder)
}

/** Starts one go module command in the dialog's folder. */
export const runModuleAction = async (
  bridge: Bridge,
  action: ModuleAction,
  argument = ''
): Promise<void> => {
  moduleTask.set({ action, error: null })
  try {
    await start(bridge, action, get(modulesFolder), argument.trim())
  } catch (failure) {
    const error = failure instanceof Error ? failure.message : String(failure)
    moduleTask.set({ action, error })
  }
}
