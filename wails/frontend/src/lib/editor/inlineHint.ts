// The line decorations of the prototype: the sand debug line and the explanation under a problem.
import type { EditorState, Range } from '@codemirror/state'
import { Decoration, EditorView, WidgetType, type DecorationSet } from '@codemirror/view'
import { marksField, type ProblemMark } from './marks'

class HintWidget extends WidgetType {
  constructor(
    readonly title: string,
    readonly detail: string,
    readonly severity: string
  ) {
    super()
  }
  override eq(other: HintWidget): boolean {
    return (
      other.title === this.title && other.detail === this.detail && other.severity === this.severity
    )
  }
  override toDOM(): HTMLElement {
    const box = document.createElement('div')
    box.className = `cm-inline-hint cm-inline-hint-${this.severity}`
    box.append(this.title + ' ')
    if (this.detail !== this.title) {
      const detail = document.createElement('span')
      detail.textContent = this.detail
      box.append(detail)
    }
    return box
  }
}

const inDocument = (state: EditorState, line: number): boolean =>
  line >= 1 && line <= state.doc.lines

const hintFor = (state: EditorState, problem: ProblemMark): Range<Decoration>[] => {
  if (!inDocument(state, problem.line)) return []
  const line = state.doc.line(problem.line)
  const widget = new HintWidget(problem.title, problem.detail, problem.severity)
  return [Decoration.widget({ widget, block: true, side: 1 }).range(line.to)]
}

const buildDecorations = (state: EditorState): DecorationSet => {
  const marks = state.field(marksField)
  const ranges = marks.problems.flatMap((problem) => hintFor(state, problem))
  if (marks.debugLine !== null && inDocument(state, marks.debugLine)) {
    const line = state.doc.line(marks.debugLine)
    ranges.push(Decoration.line({ class: 'cm-debug-line' }).range(line.from))
  }
  return Decoration.set(ranges, true)
}

export const inlineHints = EditorView.decorations.compute([marksField, 'doc'], buildDecorations)
