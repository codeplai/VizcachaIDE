import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { SAMPLE_CALC, SAMPLE_MAIN } from '../bridge/mockData'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  connectStores,
  fileTree,
  findReferences,
  groupByFile,
  openTabs,
  outputTab,
  referencesState,
  revealRequest
} from '../stores'
import BottomPanel from './BottomPanel.svelte'
import ReferencesPanel from './ReferencesPanel.svelte'

const atSumar = { file: SAMPLE_MAIN, line: 5, column: 7 }

// The panel is wired to the shared mock bridge; stores get their own for the actions.
const bridge = createMockBridge().bridge

beforeAll(() => setupI18n('en'))

beforeEach(async () => {
  fileTree.set(null)
  buffers.set({})
  openTabs.set([])
  activePath.set(null)
  revealRequest.set(null)
  referencesState.set(null)
  outputTab.set('output')
  await connectStores(bridge)
})

afterEach(() => cleanup())

describe('grouping', () => {
  it('groups by file in order of appearance', () => {
    const at = (file: string, line: number) => ({
      range: {
        start: { file, line, column: 1 },
        end: { file, line, column: 2 }
      },
      preview: ''
    })
    const groups = groupByFile([at('a.go', 1), at('b.go', 2), at('a.go', 9)])
    expect(groups.map((group) => [group.file, group.items.length])).toEqual([
      ['a.go', 2],
      ['b.go', 1]
    ])
  })
})

describe('References panel', () => {
  it('invites to search before the first search', () => {
    render(ReferencesPanel)
    expect(screen.getByText(/press Shift\+F12/)).toBeTruthy()
  })

  it('lists the uses grouped by file, with the count and the text of each line', async () => {
    render(ReferencesPanel)
    await findReferences(bridge, atSumar, 'sumar')
    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('4 uses'))
    expect(screen.getByRole('status').textContent).toContain('“sumar” in 2 files')
    const list = screen.getByRole('list', { name: 'List of references' })
    expect(within(list).getByText('main.go')).toBeTruthy()
    expect(within(list).getByText('calculadora.go')).toBeTruthy()
    expect(within(list).getByText('return sumar(a, a)')).toBeTruthy()
    expect(within(list).getByText('5:6')).toBeTruthy()
  })

  it('opens the file at the position of the result that is clicked', async () => {
    render(ReferencesPanel)
    await findReferences(bridge, atSumar, 'sumar')
    const use = await screen.findByRole('button', { name: /return sumar\(a, a\)/ })
    await fireEvent.click(use)
    await waitFor(() => expect(get(activePath)).toBe(SAMPLE_CALC))
    expect(get(revealRequest)?.location).toMatchObject({ file: SAMPLE_CALC, line: 8 })
  })

  it('says so when nothing is found', async () => {
    render(ReferencesPanel)
    await findReferences(bridge, { file: SAMPLE_MAIN, line: 5, column: 2 }, 'func')
    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('No uses found'))
  })
})

describe('Bottom panel', () => {
  it('shows the References tab, with a count, and switches to it when a search starts', async () => {
    render(BottomPanel)
    await findReferences(bridge, atSumar, 'sumar')
    expect(get(outputTab)).toBe('references')
    const tab = await screen.findByRole('tab', { name: /References/ })
    expect(tab.getAttribute('aria-selected')).toBe('true')
    await waitFor(() => expect(tab.textContent).toContain('4'))
  })
})
