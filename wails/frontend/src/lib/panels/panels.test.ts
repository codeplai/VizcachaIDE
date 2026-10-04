import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge, mockControls, type Unsubscribe } from '../bridge'
import { setupI18n } from '../i18n'
import {
  activePath,
  buffers,
  collapsedFolders,
  connectStores,
  debugState,
  fileTree,
  openTabs,
  outline,
  refreshOutline,
  resetRun,
  revealRequest,
  settings,
  stopProgram
} from '../stores'
import AssistantPanel from './AssistantPanel.svelte'
import FilesPanel from './FilesPanel.svelte'
import OutlinePanel from './OutlinePanel.svelte'
import OutputPanel from './OutputPanel.svelte'
import ProblemsPanel from './ProblemsPanel.svelte'

let stop: Unsubscribe | undefined

const reset = (): void => {
  fileTree.set(null)
  activePath.set(null)
  openTabs.set([])
  buffers.set({})
  revealRequest.set(null)
  collapsedFolders.set(new Set())
}

beforeAll(() => {
  setupI18n('en')
})

beforeEach(async () => {
  reset()
  stop = await connectStores(bridge)
  await mockControls?.play('write')
  resetRun()
})

afterEach(() => {
  cleanup()
  stop?.()
})

const play = async (scenario: 'write' | 'error' | 'debug'): Promise<void> => {
  await mockControls?.play(scenario)
}

describe('Files panel', () => {
  it('shows the empty state with a button that opens a folder', async () => {
    render(FilesPanel)
    expect(screen.getByText(/No folder open/)).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Open folder' }))
    await waitFor(() => expect(get(fileTree)?.name).toBe('hola-go'))
    expect(get(settings)?.lastFolder).toBe('hola-go')
  })

  it('opens a file on double click and only selects it on a single click', async () => {
    fileTree.set(await bridge.files.openFolder())
    render(FilesPanel)
    const file = screen.getByRole('button', { name: 'main.go' })
    await fireEvent.click(file)
    expect(get(activePath)).toBeNull()
    await fireEvent.dblClick(file)
    await waitFor(() => expect(get(activePath)).toBe('hola-go/main.go'))
    expect(get(openTabs)).toEqual(['hola-go/main.go'])
  })

  it('collapses a folder', async () => {
    fileTree.set(await bridge.files.openFolder())
    render(FilesPanel)
    await fireEvent.click(screen.getByRole('button', { name: /Hide the files of hola-go/ }))
    expect(screen.queryByText('main.go')).toBeNull()
  })
})

describe('Outline panel', () => {
  it('shows its empty state, then the symbols, and navigates on click', async () => {
    const view = render(OutlinePanel)
    expect(screen.getByText(/Functions and types of this file/)).toBeTruthy()
    await refreshOutline(bridge, 'hola-go/main.go')
    await waitFor(() => expect(get(outline)).toHaveLength(2))
    await fireEvent.click(await view.findByRole('button', { name: /sumar/ }))
    await waitFor(() => expect(get(revealRequest)?.location.line).toBe(5))
  })
})

describe('Output panel', () => {
  it('prints the program output and the success line', async () => {
    render(OutputPanel)
    expect(screen.getByText(/Nothing here yet/)).toBeTruthy()
    await play('write')
    await waitFor(() => expect(screen.getByText('Hola, Go')).toBeTruthy())
    expect(screen.getByText(/Finished in 0\.4 s/)).toBeTruthy()
  })

  it('turns file:line:column into a link that navigates', async () => {
    fileTree.set(await bridge.files.openFolder())
    render(OutputPanel)
    await play('error')
    const link = await screen.findByRole('button', { name: './main.go:11:5' })
    await fireEvent.click(link)
    await waitFor(() => expect(get(revealRequest)?.location).toMatchObject({ line: 11, column: 5 }))
    expect(get(activePath)).toBe('hola-go/main.go')
    expect(screen.getByText(/there is 1 problem/)).toBeTruthy()
  })

  it('says it was stopped when the user stops the program', async () => {
    render(OutputPanel)
    await play('write')
    await stopProgram(bridge)
    await waitFor(() => expect(screen.getByText('■ Stopped.')).toBeTruthy())
  })
})

describe('Problems panel', () => {
  it('is clickable and goes to the line', async () => {
    fileTree.set(await bridge.files.openFolder())
    render(ProblemsPanel)
    expect(screen.getByText(/No problems found/)).toBeTruthy()
    await play('error')
    await fireEvent.click(await screen.findByRole('button', { name: /main\.go:11:5/ }))
    await waitFor(() => expect(get(revealRequest)?.location.line).toBe(11))
  })
})

describe('Assistant panel', () => {
  it('shows tips, then the error card with its two actions', async () => {
    render(AssistantPanel)
    expect(screen.getByText('Press Run or F5')).toBeTruthy()
    await play('error')
    expect(await screen.findByText('The variable “resultado” is never used')).toBeTruthy()
    expect(screen.getByText('Try this:')).toBeTruthy()
    expect(screen.getByText('./main.go:11:5: declared and not used: resultado')).toBeTruthy()
    const open = vi.spyOn(bridge.system, 'openUrl').mockImplementation(() => {})
    await fireEvent.click(screen.getByRole('button', { name: 'Search this error' }))
    expect(open.mock.calls[0]?.[0]).toContain('google.com/search?q=golang')
    open.mockRestore()
    await fireEvent.click(screen.getByRole('button', { name: 'Go to line 11' }))
    await waitFor(() => expect(get(revealRequest)?.location.line).toBe(11))
  })

  it('shows variables with the just-changed mark and the stack while debugging', async () => {
    render(AssistantPanel)
    await play('debug')
    expect(await screen.findByText('Variables in sumar()')).toBeTruthy()
    expect(screen.getByText('int · just changed')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: /main · main\.go:11/ }))
    await waitFor(() => expect(get(revealRequest)?.location.line).toBe(11))
    await fireEvent.click(screen.getByRole('tab', { name: 'Goroutines' }))
    expect(screen.getByText(/Goroutine 1 · main\.main/)).toBeTruthy()
  })

  it('loads the fields of a struct on demand and shows "Loading…"', async () => {
    render(AssistantPanel)
    await play('debug')
    await bridge.debug.stepOver()
    const toggle = await screen.findByRole('button', { name: /Show the contents of punto/ })
    await fireEvent.click(toggle)
    expect(screen.getByText('Loading…')).toBeTruthy()
    await waitFor(() => expect(screen.queryByText('Loading…')).toBeNull())
    expect(screen.getByText('X')).toBeTruthy()
    expect(get(debugState)?.variables.at(-1)?.children).toHaveLength(2)
    await fireEvent.click(screen.getByRole('button', { name: /Hide the contents of punto/ }))
    expect(screen.queryByText('X')).toBeNull()
  })
})
