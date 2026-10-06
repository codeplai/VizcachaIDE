import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage } from '../domain'
import { newFileTemplates } from '../editor/templates'
import { profiles } from './codeLanguages'
import { activePath, buffers, openTabs } from './files'

const UNTITLED_DIR = 'untitled/'

/** The unsaved files of this session: absolute paths the backend gives (LanguageService.UntitledFile). */
const untitledPaths = new Set<string>()

export const isUntitled = (path: string): boolean =>
  path.startsWith(UNTITLED_DIR) || untitledPaths.has(path)

const baseName = (path: string): string =>
  path.slice(Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\')) + 1)

/** The extension of a new file: the profile's first one, so the same lookup finds its language. */
const extensionFor = (codeLanguage: CodeLanguage): string =>
  get(profiles).find((profile) => profile.id === codeLanguage)?.extensions[0] ??
  newFileTemplates[codeLanguage].extension

const nextUntitledName = (codeLanguage: CodeLanguage): string => {
  const taken = new Set(Object.keys(get(buffers)).filter(isUntitled).map(baseName))
  const extension = extensionFor(codeLanguage)
  for (let number = 1; ; number++) {
    const name = `${number === 1 ? 'main' : `main${number}`}${extension}`
    if (!taken.has(name)) return name
  }
}

/**
 * Opens a file that is not on disk yet, with the language's template; running it uses
 * `RunService.RunUntitled`, which decides the language by the name's extension.
 */
export const openUntitled = async (
  bridge: Bridge,
  codeLanguage: CodeLanguage,
  program: 'blank' | 'hello' = 'blank'
): Promise<string> => {
  const path = await bridge.language.untitledFile(nextUntitledName(codeLanguage))
  untitledPaths.add(path)
  const source = newFileTemplates[codeLanguage][program]
  buffers.update((all) => ({ ...all, [path]: source }))
  await bridge.language.openDocument(path, source)
  openTabs.update((tabs) => [...tabs, path])
  activePath.set(path)
  return path
}
