import type { InlayHint, SourceRange } from '../domain'
import { SAMPLE_MAIN } from './mockData'

const hint = (line: number, column: number, label: string, kind: InlayHint['kind']): InlayHint => ({
  line,
  column,
  label,
  kind,
  paddingLeft: false,
  paddingRight: kind === 'parameter'
})

/** A few hints for the Go sample, so `npm run dev` shows them; none for other files. */
export const sampleInlayHints = (visible: SourceRange): InlayHint[] => {
  if (visible.start.file !== SAMPLE_MAIN) return []
  const all = [
    hint(6, 10, ': int', 'type'),
    hint(11, 14, ': int', 'type'),
    hint(11, 24, 'a:', 'parameter'),
    hint(11, 27, 'b:', 'parameter')
  ]
  return all.filter((item) => item.line >= visible.start.line && item.line <= visible.end.line)
}
