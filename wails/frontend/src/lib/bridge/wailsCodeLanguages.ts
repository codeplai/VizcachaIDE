// Transitional (M0): the contract v3 APIs on top of the 2.0 Go services. Track N5 adds
// bridge.CodeLanguagesService and bridge.PackagesService; then these call them directly and the
// conversion from ToolchainInfo disappears.
import * as RunService from '../../../wailsjs/go/bridge/RunService'
import type { CodeLanguage, ToolchainInfo, ToolStatus } from '../domain'
import { languageProfiles } from './languageProfiles'
import type { CodeLanguagesApi, PackagesApi } from './types'

const unsupported = (codeLanguage: CodeLanguage): Promise<never> =>
  Promise.reject(new Error(`this language does not support the action: ${codeLanguage}`))

/** The 2.0 backend only has Go tools, in the old ToolchainInfo shape. */
export const toolStatusesOf = (info: ToolchainInfo): ToolStatus[] => [
  {
    id: 'go',
    codeLanguage: 'go',
    role: 'runtime',
    version: info.goVersion,
    source: info.goSource,
    path: ''
  },
  {
    id: 'dlv',
    codeLanguage: 'go',
    role: 'debugAdapter',
    version: info.delveVersion,
    source: info.delveSource,
    path: ''
  },
  {
    id: 'gopls',
    codeLanguage: 'go',
    role: 'languageServer',
    version: info.goplsVersion,
    source: info.goplsSource,
    path: ''
  }
]

export const createCodeLanguagesApi = (): CodeLanguagesApi => ({
  profiles: async () => languageProfiles,
  tools: async () => toolStatusesOf((await RunService.Toolchain()) as unknown as ToolchainInfo)
})

export const createPackagesApi = (): PackagesApi => ({
  init: (language, dir, name) =>
    language === 'go' ? RunService.ModInit(dir, name) : unsupported(language),
  add: (language, dir, pkg) =>
    language === 'go' ? RunService.ModGet(dir, pkg) : unsupported(language),
  remove: (language) => unsupported(language),
  tidy: (language, dir) => (language === 'go' ? RunService.ModTidy(dir) : unsupported(language)),
  list: (language) => unsupported(language)
})
