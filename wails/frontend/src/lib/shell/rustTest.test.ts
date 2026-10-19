// "Test" in the title bar: it exists only for languages that can run the tests of a project.
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { bridge } from '../bridge'
import { SAMPLE_MAIN } from '../bridge/mockData'
import { languageProfiles } from '../bridge/languageProfiles'
import { RUST_MAIN } from '../bridge/mockRust'
import { setupI18n } from '../i18n'
import {
  activePath,
  connectStores,
  notice,
  openFile,
  openTabs,
  profiles,
  resetRun,
  runLines,
  testActiveFile
} from '../stores'
import TestButton from './TestButton.svelte'

beforeAll(() => setupI18n('en'))

let stop: (() => void) | undefined
beforeEach(() => profiles.set(languageProfiles))
afterEach(() => {
  cleanup()
  stop?.()
  stop = undefined
  activePath.set(null)
  openTabs.set([])
  resetRun()
})

describe('Test button', () => {
  it('runs cargo test for a Rust file and prints the report as the run', async () => {
    stop = await connectStores(bridge)
    await openFile(bridge, RUST_MAIN)
    render(TestButton)
    await fireEvent.click(await screen.findByRole('button', { name: 'Test' }))
    await waitFor(() =>
      expect(get(runLines).some((line) => line.text.includes('test result: ok. 2 passed'))).toBe(
        true
      )
    )
    expect(get(runLines)[0]?.text).toBe('cargo test')
  })

  it('is not offered for Go', async () => {
    stop = await connectStores(bridge)
    await openFile(bridge, SAMPLE_MAIN)
    render(TestButton)
    expect(screen.queryByRole('button', { name: 'Test' })).toBeNull()
  })

  it('says so when the language has no tests', async () => {
    stop = await connectStores(bridge)
    await openFile(bridge, SAMPLE_MAIN)
    await testActiveFile(bridge)
    expect(get(notice)?.messageKey).toBe('errors.unsupportedAction')
  })
})
