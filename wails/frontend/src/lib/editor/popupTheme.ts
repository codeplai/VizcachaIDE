// Search panel, tooltips and the suggestions list, in the prototype's tokens.
import { EditorView } from '@codemirror/view'

const popup = {
  background: 'var(--win)',
  color: 'var(--ink)',
  border: '1px solid var(--line)',
  borderRadius: '10px',
  boxShadow: '0 10px 24px -14px rgba(8, 18, 28, 0.55)',
  font: '400 13px/1.45 var(--ui)'
}

export const popupTheme = EditorView.theme({
  '.cm-tooltip': popup,
  '.cm-tooltip-hover, .cm-tooltip-autocomplete': { overflow: 'hidden' },
  '.cm-doc-tooltip': {
    padding: '8px 12px',
    maxWidth: '460px',
    whiteSpace: 'pre-wrap',
    font: '400 13px/1.45 var(--mono)'
  },
  '.cm-signature b': { color: 'var(--go)' },
  '.cm-tooltip.cm-completionInfo': { padding: '0', margin: '0 4px' },
  '.cm-tooltip-autocomplete > ul': { font: '400 13px/1.4 var(--mono)', maxHeight: '240px' },
  '.cm-tooltip-autocomplete > ul > li': { padding: '3px 10px' },
  '.cm-tooltip-autocomplete > ul > li[aria-selected]': {
    background: 'var(--go-soft)',
    color: 'var(--ink)'
  },
  '.cm-completionIcon-snippet:after': { content: "'✂'", fontSize: '85%' },
  '.cm-completionDetail': { color: 'var(--muted)', fontStyle: 'normal', marginLeft: '10px' },
  '.cm-tooltip-lint': { padding: '0' },
  '.cm-diagnostic': { padding: '8px 12px', font: '400 13px/1.45 var(--ui)' },
  '.cm-diagnostic-error': { borderLeft: '3px solid var(--err)' },
  '.cm-diagnostic-warning': { borderLeft: '3px solid var(--sand)' },
  '.cm-diagnostic-info': { borderLeft: '3px solid var(--go)' },
  '.cm-panels': {
    backgroundColor: 'var(--chrome)',
    color: 'var(--ink)',
    borderBottom: '1px solid var(--line)',
    font: '400 13px var(--ui)'
  },
  '.cm-panel.cm-search': { padding: '8px 12px', display: 'flex', flexWrap: 'wrap', gap: '6px' },
  '.cm-panel.cm-search br': { display: 'none' },
  '.cm-panel.cm-search label': { display: 'inline-flex', alignItems: 'center', gap: '4px' },
  '.cm-textfield': {
    background: 'var(--win)',
    color: 'var(--ink)',
    border: '1px solid var(--line)',
    borderRadius: '7px',
    padding: '4px 8px'
  },
  '.cm-textfield:focus': { outline: '2px solid var(--go)', outlineOffset: '-1px' },
  '.cm-button': {
    backgroundImage: 'none',
    background: 'var(--win)',
    color: 'var(--ink)',
    border: '1px solid var(--line)',
    borderRadius: '7px',
    padding: '4px 10px',
    cursor: 'pointer'
  },
  '.cm-button:active': { backgroundImage: 'none', background: 'var(--go-soft)' },
  '.cm-panel.cm-search [name=close]': { color: 'var(--muted)', right: '8px', top: '6px' },
  '.cm-searchMatch': { backgroundColor: 'var(--sand-soft)', outline: '1px solid var(--sand)' },
  '.cm-searchMatch-selected': { backgroundColor: 'var(--go-soft)', outline: '1px solid var(--go)' },
  '.cm-panel.cm-gotoLine': { padding: '8px 12px', display: 'flex', gap: '6px' }
})
