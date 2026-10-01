import { derived, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { Diagnostic, ExplainedDiagnostic, ServerStatus } from '../domain'

export const diagnosticsByFile = writable<Record<string, Diagnostic[]>>({})
export const explained = writable<ExplainedDiagnostic[]>([])
export const lspStatus = writable<ServerStatus>('starting')

const keyOf = (d: Diagnostic): string =>
  `${d.location?.file ?? ''}:${d.location?.line ?? 0}:${d.location?.column ?? 0}:${d.message}`

/** Joins what gopls published with what the Assistant explained, without duplicates. */
export const mergeProblems = (
  byFile: Record<string, Diagnostic[]>,
  explanations: ExplainedDiagnostic[]
): ExplainedDiagnostic[] => {
  const explainedByKey = new Map(explanations.map((item) => [keyOf(item.diagnostic), item]))
  const merged = new Map<string, ExplainedDiagnostic>()
  for (const diagnostic of Object.values(byFile).flat()) {
    const key = keyOf(diagnostic)
    merged.set(key, explainedByKey.get(key) ?? { diagnostic, explanation: null })
  }
  for (const [key, item] of explainedByKey) merged.set(key, item)
  return [...merged.values()]
}

export const problems = derived([diagnosticsByFile, explained], ([byFile, items]) =>
  mergeProblems(byFile, items)
)
export const problemCount = derived(problems, (items) => items.length)
export const firstProblem = derived(problems, (items) => items[0] ?? null)

export const connectDiagnostics = (bridge: Bridge): Unsubscribe => {
  const offs = [
    bridge.on('lsp:diagnostics', ({ path, diagnostics }) =>
      diagnosticsByFile.update((all) => ({ ...all, [path]: diagnostics }))
    ),
    bridge.on('assistant:explained', (items) => explained.set(items)),
    bridge.on('lsp:status', (status) => lspStatus.set(status))
  ]
  return () => offs.forEach((off) => off())
}
