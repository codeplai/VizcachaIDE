import { EditorState } from '@codemirror/state'
import { indentUnit } from '@codemirror/language'
import { describe, expect, it } from 'vitest'
import { cppProfile, goProfile, pythonProfile } from '../bridge/languageProfiles'
import { languageExtensionsFor } from './languageSupport'

const stateFor = (path: string, profile: Parameters<typeof languageExtensionsFor>[1]) =>
  EditorState.create({ doc: '', extensions: languageExtensionsFor(path, profile) })

describe('languageExtensionsFor', () => {
  it('indents Go with tabs', () => {
    const state = stateFor('main.go', goProfile)
    expect(state.facet(indentUnit)).toBe('\t')
    expect(state.tabSize).toBe(4)
  })

  it('indents Python and C++ with four spaces, as their profiles say', () => {
    expect(stateFor('a.py', pythonProfile).facet(indentUnit)).toBe('    ')
    expect(stateFor('a.cpp', cppProfile).facet(indentUnit)).toBe('    ')
  })

  it('keeps Go settings for a .go file before the profiles load, plain text otherwise', () => {
    expect(stateFor('main.go', null).facet(indentUnit)).toBe('\t')
    expect(stateFor('notes.txt', null).facet(indentUnit)).toBe('    ')
  })
})
