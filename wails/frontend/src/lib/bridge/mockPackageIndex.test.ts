import { describe, expect, it } from 'vitest'
import { searchDemoIndex } from './mockPackageIndex'

describe('the demo package index', () => {
  it('has fake results for Python, Go, Rust and C++', async () => {
    expect((await searchDemoIndex('python', 'numpy')).map((found) => found.name)).toContain('numpy')
    expect((await searchDemoIndex('go', 'uuid'))[0]?.name).toBe('github.com/google/uuid')
    expect((await searchDemoIndex('rust', 'rand'))[0]?.name).toBe('rand')
    expect((await searchDemoIndex('cpp', 'fmt'))[0]?.name).toBe('fmt')
  })

  it('finds nothing for an empty or unknown query', async () => {
    expect(await searchDemoIndex('python', '  ')).toEqual([])
    expect(await searchDemoIndex('python', 'zzzz')).toEqual([])
    expect(await searchDemoIndex('cpp', 'numpy')).toEqual([])
  })

  it('behaves like an unreachable index for "down"', async () => {
    await expect(searchDemoIndex('python', 'down')).rejects.toThrow()
  })
})
