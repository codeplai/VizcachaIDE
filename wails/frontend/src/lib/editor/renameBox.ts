// The little input that opens at the symbol when renaming (F2): prefilled and selected, Enter
// renames, Escape cancels. It is a tooltip of the editor, so it sits under the name it renames.
import { StateEffect, StateField, type Extension } from '@codemirror/state'
import { EditorView, showTooltip, type Tooltip } from '@codemirror/view'

export interface RenameBoxSpec {
  /** Where the symbol starts: the box opens here. */
  from: number
  /** The current name, prefilled and selected. */
  name: string
  /** Called with the typed name (trimmed, not empty) when the user presses Enter. */
  onSubmit: (name: string) => void
}

const setBox = StateEffect.define<Tooltip | null>()

const boxField = StateField.define<Tooltip | null>({
  create: () => null,
  update: (value, transaction) => {
    for (const effect of transaction.effects) if (effect.is(setBox)) return effect.value
    return transaction.docChanged ? null : value // the symbol moved or changed: the box is stale
  },
  provide: (field) => showTooltip.from(field)
})

const boxTheme = EditorView.theme({
  '.cm-rename': { padding: '6px 8px', display: 'grid', gap: '4px' },
  '.cm-rename input': {
    font: '600 13px var(--mono)',
    color: 'var(--ink)',
    background: 'var(--win)',
    border: '1px solid var(--go)',
    borderRadius: '6px',
    padding: '4px 8px',
    minWidth: '180px',
    outline: 'none'
  },
  '.cm-rename-hint': { font: '400 11.5px var(--ui)', color: 'var(--muted)' }
})

const buildBox = (view: EditorView, spec: RenameBoxSpec): Tooltip => {
  const dom = document.createElement('div')
  dom.className = 'cm-rename'
  const input = document.createElement('input')
  input.type = 'text'
  input.value = spec.name
  input.spellcheck = false
  input.autocomplete = 'off'
  input.setAttribute('aria-label', view.state.phrase('Rename symbol'))
  const hint = document.createElement('div')
  hint.className = 'cm-rename-hint'
  hint.textContent = view.state.phrase('Enter to rename, Esc to cancel')
  dom.append(input, hint)
  let closed = false
  const close = (refocus: boolean): void => {
    if (closed) return
    closed = true
    closeRenameBox(view)
    if (refocus) view.focus()
  }
  input.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') {
      event.preventDefault()
      close(true)
    } else if (event.key === 'Enter') {
      event.preventDefault()
      const name = input.value.trim()
      close(true)
      if (name) spec.onSubmit(name)
    }
  })
  input.addEventListener('blur', () => close(false))
  return {
    pos: spec.from,
    above: false,
    create: () => ({
      dom,
      mount: () => {
        input.focus()
        input.select()
      }
    })
  }
}

export const renameBox: Extension = [boxField, boxTheme]

export const openRenameBox = (view: EditorView, spec: RenameBoxSpec): void =>
  view.dispatch({ effects: setBox.of(buildBox(view, spec)) })

export const closeRenameBox = (view: EditorView): void => {
  if (view.state.field(boxField, false)) view.dispatch({ effects: setBox.of(null) })
}
