// The contract v3 APIs on top of the Go services bridge.CodeLanguagesService and
// bridge.PackagesService, plus the Transitional (M0) members of RunApi and SettingsApi that
// still speak ToolchainInfo. Track N4 stops using the transitional ones; then they are deleted.
import * as CodeLanguagesService from '../../../wailsjs/go/bridge/CodeLanguagesService'
import * as PackagesService from '../../../wailsjs/go/bridge/PackagesService'
import type { ToolchainInfo, ToolStatus } from '../domain'
import type { CodeLanguagesApi, PackagesApi } from './types'

// The generated classes and our plain interfaces describe the same JSON.
const fromWire = <T>(value: unknown): T => value as T
const toWire = (value: unknown): never => value as never

export const createCodeLanguagesApi = (): CodeLanguagesApi => ({
  profiles: async () => fromWire(await CodeLanguagesService.Profiles()),
  tools: async () => fromWire(await CodeLanguagesService.Tools())
})

export const createPackagesApi = (): PackagesApi => ({
  init: (language, dir, name) => PackagesService.Init(toWire(language), dir, name),
  add: (language, dir, pkg) => PackagesService.Add(toWire(language), dir, pkg),
  remove: (language, dir, pkg) => PackagesService.Remove(toWire(language), dir, pkg),
  tidy: (language, dir) => PackagesService.Tidy(toWire(language), dir),
  list: (language, dir) => PackagesService.List(toWire(language), dir)
})

/** Transitional (M0): the Go tools of ToolStatus[] in the old ToolchainInfo shape. */
export const toolchainInfoOf = (statuses: ToolStatus[]): ToolchainInfo => {
  const find = (id: string): ToolStatus | undefined =>
    statuses.find((status) => status.codeLanguage === 'go' && status.id === id)
  return {
    goVersion: find('go')?.version ?? '',
    delveVersion: find('dlv')?.version ?? '',
    goplsVersion: find('gopls')?.version ?? '',
    goSource: find('go')?.source ?? 'missing',
    delveSource: find('dlv')?.source ?? 'missing',
    goplsSource: find('gopls')?.source ?? 'missing'
  }
}

/** Transitional (M0): run.toolchain on top of CodeLanguagesService.Tools. */
export const goToolchain = async (): Promise<ToolchainInfo> =>
  toolchainInfoOf(fromWire(await CodeLanguagesService.Tools()))

/** Transitional (M0): run.modInit, run.modGet and run.modTidy on top of PackagesService. */
export const goPackages = {
  modInit: (dir: string, modulePath: string): Promise<void> =>
    PackagesService.Init(toWire('go'), dir, modulePath),
  modGet: (dir: string, pkg: string): Promise<void> => PackagesService.Add(toWire('go'), dir, pkg),
  modTidy: (dir: string): Promise<void> => PackagesService.Tidy(toWire('go'), dir)
}
