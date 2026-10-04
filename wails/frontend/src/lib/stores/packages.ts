// The Packages dialog: the package manager of the open file's language (go mod, pip...).
import { derived, get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage, PackageAction, ProjectContext } from '../domain'
import { activeCodeLanguage } from './codeLanguages'
import { activePath, baseName, fileTree } from './files'
import { lastRunConfiguration, runLines, runResult, running } from './run'

/** The project a folder is for a language: the kind the backend reports and its marker file. */
const PROJECT_MARKERS: Partial<
  Record<CodeLanguage, { kind: ProjectContext['kind']; file: string }>
> = {
  go: { kind: 'gomod', file: 'go.mod' },
  python: { kind: 'pyproject', file: 'pyproject.toml' }
}

/** The project the dialog knows about. A package command's own run must not make it forget. */
const knownProject = writable<ProjectContext | null>(null)
/** Set when the dialog ran "init" successfully for this folder. */
const createdIn = writable<string | null>(null)

/** The last command the dialog started and the error that kept it from starting, if any. */
export const packageTask = writable<{
  action: PackageAction
  codeLanguage: CodeLanguage
  error: string | null
} | null>(null)

lastRunConfiguration.subscribe((config) => {
  if (config?.project && config.project.kind !== 'folder') knownProject.set(config.project)
})

const parentOf = (path: string): string => path.replace(/[\\/][^\\/]*$/, '')

/** The folder the package commands run in: the open folder, else the active file's folder. */
export const packagesFolder = derived(
  [fileTree, activePath, lastRunConfiguration],
  ([tree, active, last]) => {
    if (tree?.path) return tree.path
    if (active) return parentOf(active)
    return last?.workingDir ?? ''
  }
)

/** True when the folder has the language's project file (as far as the IDE has seen). */
export const hasProject = derived(
  [packagesFolder, activeCodeLanguage, fileTree, knownProject, createdIn],
  ([folder, codeLanguage, tree, known, created]) => {
    const marker = PROJECT_MARKERS[codeLanguage]
    if (!folder || !marker) return false
    if (created === folder) return true
    if (known?.kind === marker.kind && known.root === folder) return true
    return tree?.path === folder && tree.children.some((node) => node.name === marker.file)
  }
)

/** Suggested project name: the folder's name, made safe for a module path. */
export const suggestedProjectName = derived(packagesFolder, (folder) =>
  baseName(folder)
    .toLowerCase()
    .replace(/[^a-z0-9._/-]+/g, '-')
    .replace(/^[-.]+|-+$/g, '')
)

const PROJECT_NAME = /^[A-Za-z0-9][A-Za-z0-9._~/-]*$/
const PACKAGE_NAMES: Partial<Record<CodeLanguage, RegExp>> = {
  go: /^[A-Za-z0-9][A-Za-z0-9._~/-]*(@[A-Za-z0-9._+-]+)?$/,
  python: /^[A-Za-z0-9][A-Za-z0-9._-]*(\[[A-Za-z0-9,_-]+\])?([=<>~!]=?[A-Za-z0-9._*+-]+)?$/
}

/** True when `init` can take the text as a project name (a Go module path). */
export const isValidProjectName = (name: string): boolean => PROJECT_NAME.test(name.trim())

/** True when the language's package manager can take the text as a package (with its version). */
export const isValidPackage = (name: string, codeLanguage: CodeLanguage = 'go'): boolean =>
  (PACKAGE_NAMES[codeLanguage] ?? PACKAGE_NAMES.go)?.test(name.trim()) ?? false

/** Output of the command: the lines the run store collected, without the "start" marker. */
export const packageOutput = derived(runLines, (lines) =>
  lines.filter((line) => line.kind !== 'start').map((line) => line.text)
)

/** True while a command started from the dialog is running. */
export const packageBusy = derived(
  [running, packageTask],
  ([isRunning, task]) => isRunning && task !== null
)

// A successful "init" means the folder now has its project file.
runResult.subscribe((result) => {
  if (get(packageTask)?.action === 'init' && result?.exitCode === 0) {
    createdIn.set(get(packagesFolder))
  }
})

const start = (
  bridge: Bridge,
  action: PackageAction,
  codeLanguage: CodeLanguage,
  folder: string,
  argument: string
): Promise<void> => {
  if (action === 'init') return bridge.packages.init(codeLanguage, folder, argument)
  if (action === 'add') return bridge.packages.add(codeLanguage, folder, argument)
  if (action === 'remove') return bridge.packages.remove(codeLanguage, folder, argument)
  if (action === 'tidy') return bridge.packages.tidy(codeLanguage, folder)
  return bridge.packages.list(codeLanguage, folder)
}

/** Starts one package command of the active language in the dialog's folder. */
export const runPackageAction = async (
  bridge: Bridge,
  action: PackageAction,
  argument = ''
): Promise<void> => {
  const codeLanguage = get(activeCodeLanguage)
  packageTask.set({ action, codeLanguage, error: null })
  try {
    await start(bridge, action, codeLanguage, get(packagesFolder), argument.trim())
  } catch (failure) {
    const error = failure instanceof Error ? failure.message : String(failure)
    packageTask.set({ action, codeLanguage, error })
  }
}
