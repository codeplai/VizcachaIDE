export interface Segment {
  text: string
  matched: boolean
}

/** Splits `text` so the letters at `indices` (positions of `text` shifted by `offset`) stand apart. */
export const highlightSegments = (text: string, indices: number[], offset = 0): Segment[] => {
  const hit = new Set(indices.map((index) => index - offset))
  const segments: Segment[] = []
  for (let i = 0; i < text.length; i++) {
    const matched = hit.has(i)
    const last = segments[segments.length - 1]
    if (last && last.matched === matched) last.text += text[i]
    else segments.push({ text: text[i] ?? '', matched })
  }
  return segments
}
