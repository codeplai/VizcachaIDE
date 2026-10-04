import { language as languageFacet } from '@codemirror/language'
import { EditorState } from '@codemirror/state'
import { get } from 'svelte/store'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { languageProfiles, pythonProfile } from '../bridge/languageProfiles'
import { languageExtensionsFor } from '../editor/languageSupport'
import { newFileTemplates } from '../editor/templates'
import {
  debugActive,
  debuggedPath,
  debugInputEnabled,
  debugOutput,
  debugStarting,
  enabledProfiles,
  lastRunConfiguration,
  outputLines,
  profiles,
  programInputOpen,
  resetRun,
  runLines,
  running,
  sendProgramInput,
  settings,
  startingCodeLanguage,
  toggledLanguages,
  toggleEnabledLanguage
} from '.'

const ids = () => get(enabledProfiles).map((profile) => profile.id)

describe('enabled languages', () => {
  const { bridge } = createMockBridge()

  beforeEach(async () => {
    profiles.set(languageProfiles)
    settings.set(await bridge.settings.get())
  })

  it('shows every profile while the student has not chosen', () => {
    expect(ids()).toEqual(['go', 'python', 'cpp', 'rust'])
  })

  it('shows only the chosen ones, in profile order', () => {
    settings.update((current) =>
      current ? { ...current, enabledCodeLanguages: ['python', 'go'] } : current
    )
    expect(ids()).toEqual(['go', 'python'])
  })

  it('keeps at least one language ticked', () => {
    const all = ['go', 'python', 'cpp', 'rust'] as const
    expect(toggledLanguages([...all], ['go'], 'go')).toBeNull()
    expect(toggledLanguages([...all], ['go', 'python'], 'go')).toEqual(['python'])
  })

  it('saves the empty list (all) when every language is ticked again', () => {
    const all = ['go', 'python', 'cpp', 'rust'] as const
    expect(toggledLanguages([...all], ['go', 'python', 'cpp'], 'rust')).toEqual([])
  })

  it('saves the choice and moves the default language off one that was cleared', async () => {
    const { bridge: own } = createMockBridge()
    settings.set(await own.settings.get())
    const save = vi.spyOn(own.settings, 'save')
    await toggleEnabledLanguage(own, 'go')
    expect(save).toHaveBeenCalledWith(
      expect.objectContaining({
        enabledCodeLanguages: ['python', 'cpp', 'rust'],
        defaultCodeLanguage: 'python'
      })
    )
  })

  it('starts new files in the default language if enabled, else in the first enabled one', () => {
    settings.update((current) =>
      current ? { ...current, enabledCodeLanguages: ['python'] } : current
    )
    expect(startingCodeLanguage('go')).toBe('python')
    expect(startingCodeLanguage('python')).toBe('python')
  })
})

describe('Python templates', () => {
  it('has the greeting and a blank file', () => {
    expect(newFileTemplates.python.hello).toBe('print("Hola, Python")\n')
    expect(newFileTemplates.python.blank).toBe('')
  })
})

describe('Python syntax', () => {
  it('highlights .py files with the Python language and four-space indentation', () => {
    const state = EditorState.create({
      doc: 'def f():\n    pass\n',
      extensions: languageExtensionsFor('app.py', pythonProfile)
    })
    expect(state.facet(languageFacet)?.name).toBe('python')
    expect(state.tabSize).toBe(4)
  })
})

describe('debug input', () => {
  const start = (path: string) => {
    debuggedPath.set(path)
    debugActive.set(true)
  }

  beforeEach(() => {
    profiles.set(languageProfiles)
    resetRun()
    debugOutput.set([])
    lastRunConfiguration.set(null)
  })
  afterEach(() => {
    debugActive.set(false)
    debugStarting.set(false)
    debuggedPath.set(null)
    debugOutput.set([])
  })

  it('is enabled for a Python file and disabled for a Go one', () => {
    start('app.py')
    expect(get(debugInputEnabled)).toBe(true)
    expect(get(programInputOpen)).toBe(true)
    start('main.go')
    expect(get(debugInputEnabled)).toBe(false)
    expect(get(programInputOpen)).toBe(false)
  })

  it('is closed when nothing runs', () => {
    expect(get(programInputOpen)).toBe(false)
    running.set(true)
    expect(get(programInputOpen)).toBe(true)
    running.set(false)
  })

  it('keeps the keyboard notice for Go only', () => {
    start('main.go')
    expect(get(outputLines).some((line) => line.key === 'run.debugStdin')).toBe(true)
    start('app.py')
    expect(get(outputLines).some((line) => line.key === 'run.debugStdin')).toBe(false)
  })

  it('shows the program text of the debuggee, a prompt and its answer on one line', () => {
    start('app.py')
    debugOutput.set([
      { text: 'Starting', category: 'console' },
      { text: 'Name: ', category: 'stdout' },
      { text: 'Ada\r\n', category: 'stdout' },
      { text: 'Hello Ada\n', category: 'stdout' }
    ])
    const texts = get(outputLines)
      .filter((line) => line.text !== undefined)
      .map((line) => line.text)
    expect(texts).toEqual(['Name: Ada', 'Hello Ada'])
  })

  it('keeps what the debuggee printed after the session ends, until the next run', () => {
    start('app.py')
    debugOutput.set([{ text: 'Te gusta el verde\n', category: 'stdout' }])
    debugStarting.set(false)
    debugActive.set(false)
    const ended = get(outputLines)
    expect(ended[0]?.key).toBe('run.debugEnded')
    expect(ended.map((line) => line.text)).toContain('Te gusta el verde')
    debugOutput.set([])
    expect(get(outputLines).some((line) => line.key === 'run.debugEnded')).toBe(false)
  })

  it('sends the typed line without repeating it: the terminal echoes it', async () => {
    const { bridge } = createMockBridge()
    const write = vi.spyOn(bridge.run, 'writeInput')
    start('app.py')
    await sendProgramInput(bridge, 'Ada')
    expect(write).toHaveBeenCalledWith('Ada')
    expect(get(runLines)).toEqual([])
  })
})
