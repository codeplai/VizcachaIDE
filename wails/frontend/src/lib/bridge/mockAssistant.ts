import type { Diagnostic, ExplainedDiagnostic } from '../domain'
import type { Language } from '../language'
import { sampleDiagnostic, sampleExplanation } from './mockData'
import type { Emit, Scenario } from './mockScenarios'
import type { AssistantApi } from './types'

/** Like the real backend: explains what it knows, emits `assistant:explained` and returns the list. */
export const mockAssistant = (
  scenario: () => Scenario,
  language: () => Language,
  emit: Emit
): AssistantApi => {
  const known = sampleDiagnostic().message
  const answer = (diagnostics: Diagnostic[]): ExplainedDiagnostic[] => {
    const items = diagnostics.map((diagnostic) => ({
      diagnostic,
      explanation: diagnostic.message === known ? sampleExplanation(language()) : null
    }))
    emit('assistant:explained', items)
    return items
  }
  return {
    explain: async () => answer(scenario() === 'error' ? [sampleDiagnostic()] : []),
    explainDiagnostics: async (_codeLanguage, diagnostics) => answer(diagnostics)
  }
}
