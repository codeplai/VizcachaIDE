// Parameter help shown above the cursor after typing "(" or ",".
import { StateEffect, StateField, type Extension, type Transaction } from '@codemirror/state'
import { EditorView, keymap, showTooltip, type Tooltip } from '@codemirror/view'
import type { SignatureHelp } from '../domain'
import { locationAt, type DocumentContext, type LanguageApi } from './documentContext'

interface SignatureAt {
  pos: number
  help: SignatureHelp
}

const showSignature = StateEffect.define<SignatureAt | null>()

/** Builds the tooltip body, with the parameter being typed in bold. */
export const renderSignature = (help: SignatureHelp): HTMLElement => {
  const box = document.createElement('div')
  box.className = 'cm-doc-tooltip cm-signature'
  const active = help.parameters[help.activeParameter]
  const at = active ? help.label.indexOf(active) : -1
  if (!active || at < 0) {
    box.textContent = help.label
    return box
  }
  const bold = document.createElement('b')
  bold.textContent = active
  box.append(help.label.slice(0, at), bold, help.label.slice(at + active.length))
  return box
}

const toTooltip = ({ pos, help }: SignatureAt): Tooltip => ({
  pos,
  above: true,
  create: () => ({ dom: renderSignature(help) })
})

/** The help goes away when the cursor leaves the call: it moves before the "(" or past a ")". */
const leftTheCall = (transaction: Transaction, pos: number): boolean => {
  if (!transaction.selection) return false
  const head = transaction.newSelection.main.head
  return head < pos || transaction.state.sliceDoc(head - 1, head) === ')'
}

const signatureField = StateField.define<Tooltip | null>({
  create: () => null,
  update: (value, transaction) => {
    for (const effect of transaction.effects) {
      if (effect.is(showSignature)) return effect.value ? toTooltip(effect.value) : null
    }
    if (!value) return null
    const pos = transaction.changes.mapPos(value.pos)
    return leftTheCall(transaction, pos) ? null : { ...value, pos }
  },
  provide: (field) => showTooltip.from(field)
})

const typedCallCharacter = (transaction: Transaction): boolean => {
  if (!transaction.isUserEvent('input.type')) return false
  let typed = false
  transaction.changes.iterChanges((_fa, _ta, _fb, _tb, inserted) => {
    const text = inserted.toString()
    if (text.startsWith('(') || text === ',') typed = true
  })
  return typed
}

const closeSignature = (view: EditorView): boolean => {
  if (!view.state.field(signatureField)) return false
  view.dispatch({ effects: showSignature.of(null) })
  return true
}

export const signatureHelp = (language: LanguageApi, file: DocumentContext): Extension => [
  signatureField,
  keymap.of([{ key: 'Escape', run: closeSignature }]),
  EditorView.updateListener.of((update) => {
    if (!update.docChanged || !update.transactions.some(typedCallCharacter)) return
    const path = file.path()
    if (!path) return
    const view = update.view
    const pos = update.state.selection.main.head
    void file
      .flush()
      .then(() => language.signatureHelp(locationAt(update.state, pos, path)))
      .then((help) => {
        if (!help || view.state.selection.main.head < pos) return
        view.dispatch({ effects: showSignature.of({ pos, help }) })
      })
  })
]
