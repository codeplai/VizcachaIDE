import { describe, expect, it } from 'vitest'
import { createMockBridge } from './mock'

describe('mock projects service', () => {
  it('creates the template in the mock file system', async () => {
    const { bridge } = createMockBridge()
    const location = await bridge.files.chooseFolder()
    const project = await bridge.projects.create('rust', location, 'Mi Proyecto')
    expect(project.mainFile).toBe(`${location}/Mi Proyecto/src/main.rs`)
    const tree = await bridge.files.listTree(project.root)
    expect(tree.children.map((node) => node.name).sort()).toEqual([
      '.gitignore',
      'Cargo.toml',
      'src'
    ])
    expect(await bridge.files.readFile(project.mainFile)).toContain('read_line')
  })

  it('answers the same keyed errors as the backend', async () => {
    const { bridge } = createMockBridge()
    const location = await bridge.files.chooseFolder()
    await bridge.projects.create('python', location, 'uno')
    await expect(bridge.projects.create('python', location, 'uno')).rejects.toThrow(
      'project.errorExists'
    )
    await expect(bridge.projects.create('go', location, ' ')).rejects.toThrow(
      'project.errorNameEmpty'
    )
    await expect(bridge.projects.create('go', location, 'a|b')).rejects.toThrow(
      'project.errorNameInvalid'
    )
  })
})
