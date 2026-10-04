import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage } from '../domain'
import { newFileTemplates } from '../editor/templates'
import { profiles } from './codeLanguages'
import { activePath, buffers, openTabs } from './files'

const UNTITLED_DIR = 'untitled/'

export const isUntitled = (path: string): boolean => path.startsWith(UNTITLED_DIR)

/** The extension of a new file: the profile's first one, so the same lookup finds its language. */
const extensionFor = (codeLanguage: CodeLanguage): string =>
  get(profiles).find((profile) => profile.id === codeLanguage)?.extensions[0] ??
  newFileTemplates[codeLanguage].extension

const nextUntitledPath = (codeLanguage: CodeLanguage): string => {
  const taken = new Set(Object.keys(get(buffers)))
  const extension = extensionFor(codeLanguage)
  for (let number = 1; ; number++) {
    const path = `${UNTITLED_DIR}${number === 1 ? 'main' : `main${number}`}${extension}`
    if (!taken.has(path)) return path
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
  const path = nextUntitledPath(codeLanguage)
  const source = newFileTemplates[codeLanguage][program]
  buffers.update((all) => ({ ...all, [path]: source }))
  await bridge.language.openDocument(path, source)
  openTabs.update((tabs) => [...tabs, path])
  activePath.set(path)
  return path
}
