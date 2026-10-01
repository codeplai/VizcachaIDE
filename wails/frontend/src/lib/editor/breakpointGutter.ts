import { GutterMarker, gutter } from '@codemirror/view'
import { marksField, setMarks } from './marks'

class BreakpointMarker extends GutterMarker {
  override toDOM(): HTMLElement {
    const dot = document.createElement('span')
    dot.className = 'cm-bp'
    return dot
  }
}

const breakpointMarker = new BreakpointMarker()

/** Gutter column with the breakpoint dots; clicking a row asks the caller to toggle it. */
export const breakpointGutter = (onToggle: (line: number) => void) =>
  gutter({
    class: 'cm-bpgutter',
    lineMarker: (view, block) => {
      const number = view.state.doc.lineAt(block.from).number
      return view.state.field(marksField).breakpoints.includes(number) ? breakpointMarker : null
    },
    lineMarkerChange: (update) =>
      update.transactions.some((tr) => tr.effects.some((effect) => effect.is(setMarks))),
    initialSpacer: () => breakpointMarker,
    domEventHandlers: {
      mousedown: (view, block) => {
        onToggle(view.state.doc.lineAt(block.from).number)
        return true
      }
    }
  })
