// The editor split: a second editor beside or below the first, both on the same open files.
import { cleanup, fireEvent, render, waitFor } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_MAIN } from '../bridge/mockData'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  closeSplit,
  editorSplit,
  openFile,
  openTabs,
  splitEditor
} from '../stores'
import { get } from 'svelte/store'
import EditorPane from './EditorPane.svelte'

beforeAll(() => setupI18n('en'))
beforeEach(() => {
  openTabs.set([])
  activePath.set(null)
  buffers.set({})
  closeSplit()
})
afterEach(() => cleanup())

describe('editor split', () => {
  it('does nothing without an open file', () => {
    splitEditor('right')
    expect(get(editorSplit)).toBeNull()
  })

  it('shows a second editor of the open file and closes it', async () => {
    const { bridge } = createMockBridge()
    const view = render(EditorPane)
    await openFile(bridge, SAMPLE_MAIN)
    splitEditor('down')
    await waitFor(() => expect(view.container.querySelectorAll('.cm-editor')).toHaveLength(2))
    expect(view.container.querySelector('.code.stacked')).not.toBeNull()
    await fireEvent.click(view.getByLabelText('Close the split'))
    await waitFor(() => expect(view.container.querySelectorAll('.cm-editor')).toHaveLength(1))
    expect(get(editorSplit)).toBeNull()
  })

  it('goes away with its file', async () => {
    const { bridge } = createMockBridge()
    await openFile(bridge, SAMPLE_MAIN)
    splitEditor('left')
    expect(get(editorSplit)?.direction).toBe('left')
    buffers.set({})
    expect(get(editorSplit)).toBeNull()
    buffers.set({ [SAMPLE_MAIN]: '' })
    expect(get(editorSplit)).toBeNull()
  })
})
