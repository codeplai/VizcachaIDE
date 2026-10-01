import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { EditorView } from '@codemirror/view'
import { tags } from '@lezer/highlight'

/** Editor chrome, colored only with the CSS tokens of the prototype (light and dark for free). */
export const editorTheme = EditorView.theme({
  '&': { height: '100%', backgroundColor: 'var(--win)', color: 'var(--ink)' },
  '.cm-scroller': {
    font: '400 var(--editor-font-size, 14px)/1.75 var(--mono)',
    overflow: 'auto'
  },
  '.cm-content': { padding: '10px 0', caretColor: 'var(--ink)' },
  '.cm-line': { padding: '0 12px 0 0' },
  '&.cm-focused': { outline: 'none' },
  '.cm-cursor': { borderLeftColor: 'var(--ink)' },
  '.cm-gutters': { backgroundColor: 'var(--win)', border: 'none', color: 'var(--code-com)' },
  '.cm-lineNumbers .cm-gutterElement': {
    padding: '0 14px 0 0',
    fontSize: '0.9em',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end'
  },
  '.cm-bpgutter': { width: '22px', cursor: 'pointer' },
  '.cm-bpgutter .cm-gutterElement': { display: 'grid', placeItems: 'center' },
  '.cm-bp': { width: '10px', height: '10px', borderRadius: '50%', background: 'var(--err)' },
  '.cm-debug-line': { backgroundColor: 'var(--sand-soft)', boxShadow: 'inset 3px 0 0 var(--sand)' },
  '.cm-lintRange': {
    backgroundImage: 'none',
    paddingBottom: '0',
    textDecorationLine: 'underline',
    textDecorationStyle: 'wavy',
    textDecorationThickness: '1.5px',
    textUnderlineOffset: '4px'
  },
  '.cm-lintRange-error': { textDecorationColor: 'var(--err)' },
  '.cm-lintRange-warning': { textDecorationColor: 'var(--sand)' },
  '.cm-lintRange-info': { textDecorationColor: 'var(--go)' },
  '.cm-inline-hint': {
    margin: '2px 12px 6px 56px',
    padding: '8px 12px',
    borderRadius: '8px',
    background: 'var(--err-soft)',
    borderLeft: '3px solid var(--err)',
    font: '600 13px/1.45 var(--ui)'
  },
  '.cm-inline-hint-warning': { background: 'var(--sand-soft)', borderLeftColor: 'var(--sand)' },
  '.cm-inline-hint-info, .cm-inline-hint-hint': {
    background: 'var(--go-soft)',
    borderLeftColor: 'var(--go)'
  },
  '.cm-inline-hint span': { fontWeight: '400', color: 'var(--muted)' },
  '.cm-activeLine': { backgroundColor: 'transparent' },
  '&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionBackground':
    { backgroundColor: 'var(--go-soft)' },
  '.cm-matchingBracket': { backgroundColor: 'var(--go-soft)', outline: '1px solid var(--go)' }
})

const highlightStyle = HighlightStyle.define([
  { tag: [tags.keyword, tags.typeName], color: 'var(--code-kw)' },
  { tag: [tags.string, tags.special(tags.string)], color: 'var(--code-str)' },
  { tag: [tags.number, tags.bool, tags.null], color: 'var(--code-num)' },
  {
    tag: [tags.comment, tags.lineComment, tags.blockComment],
    color: 'var(--code-com)',
    fontStyle: 'italic'
  },
  {
    tag: [
      tags.function(tags.variableName),
      tags.definition(tags.variableName),
      tags.function(tags.propertyName)
    ],
    color: 'var(--code-fn)'
  }
])

export const editorHighlighting = syntaxHighlighting(highlightStyle)
