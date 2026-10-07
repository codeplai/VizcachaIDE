import { describe, expect, it } from 'vitest'
import type { FileNode } from '../domain'
import { highlightSegments } from './highlight'
import { fuzzyMatch } from './fuzzy'
import { rankFiles } from './quickOpenRank'
import { projectFiles, relativePath } from './skip'

const ROOT = 'C:\\work\\app'
const abs = (relative: string): string => `${ROOT}\\${relative.replace(/\//g, '\\')}`
const PATHS = [
  'main.go',
  'internal/app/search.go',
  'internal/app/replace.go',
  'docs/readme.md',
  'frontend/src/Search.svelte'
].map(abs)

describe('fuzzyMatch', () => {
  it('needs the letters in order, in any case', () => {
    expect(fuzzyMatch('mgo', 'main.go')?.indices).toEqual([0, 5, 6])
    expect(fuzzyMatch('og', 'main.go')).toBeNull()
    expect(fuzzyMatch('MAIN', 'main.go')).not.toBeNull()
  })

  it('prefers consecutive letters and word starts', () => {
    const tight = fuzzyMatch('sea', 'internal/app/search.go')
    const loose = fuzzyMatch('sea', 'src/extra/alpha.go')
    expect(tight && loose && tight.score > loose.score).toBe(true)
  })
})

describe('rankFiles', () => {
  const names = (query: string, recents: string[] = []): string[] =>
    rankFiles(ROOT, PATHS, query, recents).map((item) => item.relative)

  it('shows recent files first when the query is empty, then the rest by path', () => {
    const order = names('', [abs('docs/readme.md'), abs('main.go')])
    expect(order.slice(0, 2)).toEqual(['docs/readme.md', 'main.go'])
    expect(order.slice(2)).toEqual([...order.slice(2)].sort())
  })

  it('ranks the file name above a path match', () => {
    expect(names('search')[0]).toBe('internal/app/search.go')
    expect(names('search')).toContain('frontend/src/Search.svelte')
  })

  it('drops files that do not match and gives relative paths with the matched letters', () => {
    const [item] = rankFiles(ROOT, PATHS, 'rdm', [])
    expect(item?.relative).toBe('docs/readme.md')
    expect(item?.name).toBe('readme.md')
    expect(item?.dir).toBe('docs')
    expect(item?.indices).toHaveLength(3)
    expect(names('zzzz')).toEqual([])
  })

  it('limits the list', () => {
    const many = Array.from({ length: 200 }, (_, i) => abs(`f${i}.go`))
    expect(rankFiles(ROOT, many, '', [], 50)).toHaveLength(50)
  })
})

describe('project files', () => {
  const node = (name: string, children?: FileNode[]): FileNode => ({
    name,
    path: `/r/${name}`,
    isDir: children !== undefined,
    children: children ?? []
  })

  it('skips build output, dependencies and version control', () => {
    const tree = node('r', [
      node('a.go'),
      node('build', [node('out.txt')]),
      node('target', [node('x.rs')]),
      node('node_modules', [node('m.js')]),
      node('.git', [node('HEAD')]),
      node('src', [node('b.go')])
    ])
    expect(projectFiles(tree).map((file) => file.name)).toEqual(['a.go', 'b.go'])
  })

  it('makes paths relative with slashes', () => {
    expect(relativePath(ROOT, abs('a/b.go'))).toBe('a/b.go')
  })
})

describe('highlightSegments', () => {
  it('splits the matched letters', () => {
    expect(highlightSegments('readme', [0, 1])).toEqual([
      { text: 're', matched: true },
      { text: 'adme', matched: false }
    ])
    expect(highlightSegments('me', [4, 5], 4)).toEqual([{ text: 'me', matched: true }])
  })
})
