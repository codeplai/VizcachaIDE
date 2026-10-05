import { get } from 'svelte/store'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { languageProfiles } from '../bridge/languageProfiles'
import {
  activePath,
  fileTree,
  hasProject,
  isValidPackage,
  isValidProjectName,
  lastRunConfiguration,
  packageTask,
  packagesFolder,
  profiles,
  runPackageAction,
  suggestedProjectName
} from '.'

const { bridge } = createMockBridge()

beforeEach(() => {
  profiles.set(languageProfiles)
  packageTask.set(null)
  lastRunConfiguration.set(null)
  fileTree.set(null)
  activePath.set('C:/work/My Project/main.go')
})

describe('package commands', () => {
  it('runs the verb on the language of the open file', async () => {
    const add = vi.spyOn(bridge.packages, 'add')
    await runPackageAction(bridge, 'add', ' github.com/user/pkg ')
    expect(add).toHaveBeenCalledWith('go', 'C:/work/My Project', 'github.com/user/pkg')
    activePath.set('C:/work/app/main.py')
    const remove = vi.spyOn(bridge.packages, 'remove')
    await runPackageAction(bridge, 'remove', 'requests')
    expect(remove).toHaveBeenCalledWith('python', 'C:/work/app', 'requests')
    const list = vi.spyOn(bridge.packages, 'list')
    await runPackageAction(bridge, 'list')
    expect(list).toHaveBeenCalledWith('python', 'C:/work/app')
  })

  it('keeps the error of a command that could not start', async () => {
    vi.spyOn(bridge.packages, 'tidy').mockRejectedValue(new Error('busy'))
    await runPackageAction(bridge, 'tidy')
    expect(get(packageTask)).toMatchObject({ action: 'tidy', codeLanguage: 'go', error: 'busy' })
  })

  it('uses the open folder, and knows its project file per language', () => {
    fileTree.set({
      name: 'proj',
      path: 'C:/work/proj',
      isDir: true,
      children: [
        { name: 'pyproject.toml', path: 'C:/work/proj/pyproject.toml', isDir: false, children: [] }
      ]
    })
    expect(get(packagesFolder)).toBe('C:/work/proj')
    expect(get(hasProject)).toBe(false)
    activePath.set('C:/work/proj/app.py')
    expect(get(hasProject)).toBe(true)
  })

  it('suggests a safe project name from the folder', () => {
    expect(get(suggestedProjectName)).toBe('my-project')
  })

  it('validates names per language', () => {
    expect(isValidProjectName('example.com/hola')).toBe(true)
    expect(isValidProjectName('two words')).toBe(false)
    expect(isValidPackage('github.com/user/pkg@v1.2.3')).toBe(true)
    expect(isValidPackage('requests==2.32.0', 'python')).toBe(true)
    expect(isValidPackage('fmt', 'cpp')).toBe(true)
    expect(isValidPackage('boost-asio', 'cpp')).toBe(true)
    expect(isValidPackage('Fmt', 'cpp')).toBe(false)
    expect(isValidPackage('github.com/a/b', 'cpp')).toBe(false)
    expect(isValidPackage('requests==2.32.0', 'go')).toBe(false)
  })
})
