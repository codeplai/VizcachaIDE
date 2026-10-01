import { get } from 'svelte/store'
import { describe, expect, it } from 'vitest'
import { createMockBridge } from '../bridge/mock'
import { DEFAULT_FONT_SIZE, MAX_FONT_SIZE, MIN_FONT_SIZE, connectStores, settings } from '../stores'
import { minimalChange } from './textChange'
import { editorFontSize, nextFontSize, zoomEditor } from './zoom'
import { editorPhrases, phrasesFor } from './phrases'
import { EditorState } from '@codemirror/state'

describe('minimalChange', () => {
  it('returns null when nothing changed', () => {
    expect(minimalChange('abc', 'abc')).toBeNull()
  })

  it('keeps the common start and end', () => {
    const change = minimalChange('a := 1\nb := 2\n', 'a := 1\nb := 20\n')
    expect(change).toEqual({ from: 13, to: 13, insert: '0' })
  })

  it('handles replacing everything and empty documents', () => {
    expect(minimalChange('', 'x')).toEqual({ from: 0, to: 0, insert: 'x' })
    expect(minimalChange('abc', '')).toEqual({ from: 0, to: 3, insert: '' })
  })

  it('turns spaces into tabs in a small span (what gofmt does on save)', () => {
    const change = minimalChange('func f() {\n    x()\n}', 'func f() {\n\tx()\n}')
    expect(change).toEqual({ from: 11, to: 15, insert: '\t' })
  })
})

describe('zoom', () => {
  it('steps by one pixel inside the limits', () => {
    expect(nextFontSize(14, 'in')).toBe(15)
    expect(nextFontSize(14, 'out')).toBe(13)
    expect(nextFontSize(MAX_FONT_SIZE, 'in')).toBe(MAX_FONT_SIZE)
    expect(nextFontSize(MIN_FONT_SIZE, 'out')).toBe(MIN_FONT_SIZE)
    expect(nextFontSize(20, 'reset')).toBe(DEFAULT_FONT_SIZE)
  })

  it('saves the new size in Settings', async () => {
    const { bridge } = createMockBridge()
    const stop = await connectStores(bridge)
    await zoomEditor(bridge, 'in')
    expect(get(settings)?.fontSize).toBe(15)
    expect(get(editorFontSize)).toBe(15)
    await zoomEditor(bridge, 'reset')
    expect(get(editorFontSize)).toBe(DEFAULT_FONT_SIZE)
    stop()
  })
})

describe('phrases', () => {
  it('translates CodeMirror texts through the given function', () => {
    const phrases = phrasesFor((key) => `[${key}]`)
    expect(phrases['Find']).toBe('[editor.search.find]')
    expect(phrases['replace all']).toBe('[editor.search.replaceAll]')
    expect(phrases['Go to line']).toBe('[editor.search.goToLine]')
  })

  it('is applied as an editor state facet', () => {
    const state = EditorState.create({ extensions: editorPhrases((key) => key.toUpperCase()) })
    expect(state.phrase('Find')).toBe('EDITOR.SEARCH.FIND')
  })
})
