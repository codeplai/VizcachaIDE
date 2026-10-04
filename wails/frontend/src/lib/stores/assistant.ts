import { derived, get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { Diagnostic, RunConfiguration } from '../domain'
import { TERMINATED_BY_USER } from '../events'
import { diagnosticsByFile, problemKey, problems, sameFile } from './diagnostics'
import { activePath } from './files'
import { lastRunConfiguration, runLines, stoppedByUser } from './run'

/** The problems Go printed in the last run, as the backend parsed them (compiler, vet or panic). */
export const runDiagnostics = writable<Diagnostic[]>([])

/** How many compiler or vet problems the last run reported (a panic is not a compile failure). */
export const compileProblemCount = derived(
  runDiagnostics,
  (items) => items.filter((item) => item.source !== 'panic').length
)

/**
 * What the Assistant shows: the problems of the file the learner is looking at, plus the ones the
 * last run printed (wherever they are). Problems of other open files stay in the Problems tab.
 */
export const assistantProblems = derived(
  [problems, activePath, runDiagnostics],
  ([items, active, ran]) => {
    const fromRun = new Set(ran.map(problemKey))
    return items.filter((item) => {
      const file = item.diagnostic.location?.file
      const inActiveFile = !!active && !!file && sameFile(file, active)
      return inActiveFile || fromRun.has(problemKey(item.diagnostic))
    })
  }
)

export const assistantProblemCount = derived(assistantProblems, (items) => items.length)

const REFRESH_DELAY_MS = 300

/**
 * Asks the backend to explain everything at once: what gopls underlined plus what the last run
 * printed. The answer arrives as `assistant:explained` (texts already in the user's language).
 */
const refreshExplanations = async (bridge: Bridge): Promise<void> => {
  const lsp = Object.values(get(diagnosticsByFile)).flat()
  await bridge.assistant.explainDiagnostics([...lsp, ...get(runDiagnostics)])
}

const stderrOfRun = (): string =>
  get(runLines)
    .filter((line) => line.kind === 'stderr')
    .map((line) => line.text)
    .join('\n')

/** Counts runs, so a slow `go vet` of an older run cannot overwrite the newer one. */
let runNumber = 0

/** A "go mod ..." command also fires run events, but it has no program to vet. */
const isGoCommand = (config: RunConfiguration): boolean => /^go\s/.test(config.target)

/**
 * After a run that compiled and finished, `go vet` looks for mistakes the compiler allows (a Printf
 * with the wrong values, code that can never run). It works in the background: the run is already
 * over and the UI never waits for it. Its warnings join the run's problems.
 */
const vetInBackground = async (bridge: Bridge, config: RunConfiguration): Promise<void> => {
  const number = runNumber
  const output = await bridge.run.check(config)
  if (!output.trim() || number !== runNumber) return
  const items = await bridge.assistant.explain(config.codeLanguage, output, config.workingDir)
  if (number !== runNumber || items.length === 0) return
  runDiagnostics.set(items.map((item) => item.diagnostic))
  await refreshExplanations(bridge)
}

const explainRun = async (bridge: Bridge, exitCode: number): Promise<void> => {
  const failed = exitCode !== 0 && exitCode !== TERMINATED_BY_USER
  const config = get(lastRunConfiguration)
  const items = failed
    ? await bridge.assistant.explain(
        config?.codeLanguage ?? 'go',
        stderrOfRun(),
        config?.workingDir ?? ''
      )
    : []
  runDiagnostics.set(items.map((item) => item.diagnostic))
  await refreshExplanations(bridge)
  if (exitCode === 0 && config && !isGoCommand(config) && !get(stoppedByUser)) {
    void vetInBackground(bridge, config).catch(() => undefined)
  }
}

/** Sends Go's output and gopls' diagnostics to the Assistant so the cards are explained. */
export const connectAssistant = (bridge: Bridge): Unsubscribe => {
  let timer: ReturnType<typeof setTimeout> | undefined
  const quietly = (work: Promise<void>): void => void work.catch(() => undefined)
  const offs = [
    bridge.on('run:started', () => {
      runNumber += 1
      runDiagnostics.set([])
    }),
    bridge.on('run:finished', ({ exitCode }) => quietly(explainRun(bridge, exitCode))),
    bridge.on('lsp:diagnostics', () => {
      clearTimeout(timer)
      timer = setTimeout(() => quietly(refreshExplanations(bridge)), REFRESH_DELAY_MS)
    })
  ]
  return () => {
    clearTimeout(timer)
    offs.forEach((off) => off())
  }
}
