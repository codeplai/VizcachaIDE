// The contract v3 APIs on top of the Go services bridge.CodeLanguagesService and
// bridge.PackagesService.
import * as CodeLanguagesService from '../../../wailsjs/go/bridge/CodeLanguagesService'
import * as PackagesService from '../../../wailsjs/go/bridge/PackagesService'
import * as ProjectsService from '../../../wailsjs/go/bridge/ProjectsService'
import type { CodeLanguagesApi, PackagesApi, ProjectsApi } from './types'

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
  list: (language, dir) => PackagesService.List(toWire(language), dir),
  search: async (language, query) => fromWire(await PackagesService.Search(toWire(language), query))
})

export const createProjectsApi = (): ProjectsApi => ({
  create: async (codeLanguage, location, name) =>
    fromWire(await ProjectsService.Create(toWire(codeLanguage), location, name))
})
