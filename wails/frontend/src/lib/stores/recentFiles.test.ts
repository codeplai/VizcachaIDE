import { get } from 'svelte/store'
import { beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import type { Bridge } from '../bridge'
import {
  clearRecentFiles,
  connectStores,
  forgetRecentFile,
  notice,
  openFile,
  openRecentFile,
  rememberRecentFile,
  settings
} from '.'

let bridge: Bridge

const recent = (): string[] => get(settings)?.recentFiles ?? []

beforeEach(async () => {
  bridge = createMockBridge().bridge
  notice.set(null)
  await connectStores(bridge)
})

describe('recent files', () => {
  it('puts the newest first and removes duplicates', async () => {
    await rememberRecentFile(bridge, 'C:\\a.go')
    await rememberRecentFile(bridge, 'C:\\b.go')
    await rememberRecentFile(bridge, 'C:\\a.go')
    expect(recent()).toEqual(['C:\\a.go', 'C:\\b.go'])
  })

  it('keeps at most 10 files', async () => {
    for (let n = 1; n <= 12; n++) await rememberRecentFile(bridge, `C:\\f${n}.go`)
    expect(recent()).toHaveLength(10)
    expect(recent()[0]).toBe('C:\\f12.go')
    expect(recent()).not.toContain('C:\\f1.go')
  })

  it('forgets one file and clears the list', async () => {
    await rememberRecentFile(bridge, 'C:\\a.go')
    await rememberRecentFile(bridge, 'C:\\b.go')
    await forgetRecentFile(bridge, 'C:\\a.go')
    expect(recent()).toEqual(['C:\\b.go'])
    await clearRecentFiles(bridge)
    expect(recent()).toEqual([])
  })

  it('remembers files opened from disk', async () => {
    await openFile(bridge, 'C:\\proj\\main.go')
    expect(recent()).toEqual(['C:\\proj\\main.go'])
  })

  it('removes a missing file and tells the user', async () => {
    await rememberRecentFile(bridge, 'C:\\gone\\old.go')
    bridge.files.readFile = async () => {
      throw new Error('not found')
    }
    await openRecentFile(bridge, 'C:\\gone\\old.go')
    expect(recent()).toEqual([])
    expect(get(notice)).toMatchObject({ messageKey: 'recent.missing', values: { file: 'old.go' } })
  })
})
