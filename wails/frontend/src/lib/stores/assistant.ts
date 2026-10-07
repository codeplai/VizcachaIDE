import { derived, get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { Diagnostic, RunConfiguration } from '../domain'
import { TERMINATED_BY_USER } from '../events'
import { diagnosticsByFile, problemKey, problems, sameFile } from './diagnostics'
import { activePath } from './files'
import { lastRunConfiguration, runLines, stoppedByUser } from './run'

/** The problems Go printed in the last run, as the backend parsed them (compiler, vet or panic). */
export const runDiagnostics = writable<Diagnostic[]>([])

/** How many compiler or vet problems the last run reported (a panic or a crash is not a compile failure). */
export const compileProblemCount = derived(
  runDiagnostics,
  (items) => items.filter((item) => item.source !== 'panic' && item.source !== 'runtime').length
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
 * The backend answers each request with an `assistant:explained` event that replaces the cards.
 * Requests go one at a time, so an older answer can never arrive after a newer one.
 */
let queue: Promise<unknown> = Promise.resolve()
const inTurn = <T>(work: () => Promise<T>): Promise<T> => {
  const next = queue.then(work, work)
  queue = next.catch(() => undefined)
  return next
}

/**
 * Asks the backend to explain everything at once: what gopls underlined plus what the last run
 * printed. The answer arrives as `assistant:explained` (texts already in the user's language).
 */
const refreshExplanations = async (bridge: Bridge): Promise<void> => {
  await inTurn(() => {
    const lsp = Object.values(get(diagnosticsByFile)).flat()
    const fallback = get(lastRunConfiguration)?.codeLanguage ?? ''
    return bridge.assistant.explainDiagnostics(fallback, [...lsp, ...get(runDiagnostics)])
  })
}

/**
 * The output the explainer reads after a failed run: stderr, or everything when the program ran
 * in a terminal (a PTY merges stderr into stdout, so a Python traceback arrives as stdout).
 */
const failureOutputOfRun = (inTerminal: boolean): string =>
  get(runLines)
    .filter((line) => line.kind === 'stderr' || (inTerminal && line.kind === 'stdout'))
    .map((line) => line.text)
    .join('\n')

/** Counts runs, so a slow `go vet` of an older run cannot overwrite the newer one. */
let runNumber = 0

/**
 * A package command ("go mod tidy", "python -m pip install numpy", "cargo add rand") also fires
 * run events, but it has no program to check: its target is the command, not a file (checking it
 * made ruff report "os error 2" on a file named like the command).
 */
export const isPackageCommand = (config: RunConfiguration): boolean =>
  /^(go|cargo|python\d*(\.exe)?|py)\s/i.test(config.target)

/**
 * Whether a finished run failed and its output needs explaining. A run the user stopped did not
 * fail: Stop sends Ctrl+C first, so Python ends with a KeyboardInterrupt traceback and its own exit
 * code, which is not an error to explain.
 */
export const runFailed = (exitCode: number, stopped: boolean): boolean =>
  exitCode !== 0 && exitCode !== TERMINATED_BY_USER && !stopped

/**
 * After a run that compiled and finished, `go vet` looks for mistakes the compiler allows (a Printf
 * with the wrong values, code that can never run). It works in the background: the run is already
 * over and the UI never waits for it. Its warnings join the run's problems.
 */
const vetInBackground = async (bridge: Bridge, config: RunConfiguration): Promise<void> => {
  const number = runNumber
  const output = await bridge.run.check(config)
  if (!output.trim() || number !== runNumber) return
  const items = await inTurn(() =>
    bridge.assistant.explain(config.codeLanguage, output, config.workingDir)
  )
  if (number !== runNumber || items.length === 0) return
  runDiagnostics.set(items.map((item) => item.diagnostic))
  await refreshExplanations(bridge)
}

const explainRun = async (bridge: Bridge, exitCode: number): Promise<void> => {
  const failed = runFailed(exitCode, get(stoppedByUser))
  const config = get(lastRunConfiguration)
  const items = failed
    ? await inTurn(() =>
        bridge.assistant.explain(
          config?.codeLanguage ?? 'go',
          failureOutputOfRun(config?.echo ?? false),
          config?.workingDir ?? ''
        )
      )
    : []
  runDiagnostics.set(items.map((item) => item.diagnostic))
  await refreshExplanations(bridge)
  if (exitCode === 0 && config && !isPackageCommand(config) && !get(stoppedByUser)) {
    void vetInBackground(bridge, config).catch(() => undefined)
  }
}

/** How long output that arrives after `run:finished` waits for more before the run is explained again. */
export const LATE_OUTPUT_DELAY_MS = 150

/**
 * Sends Go's output and gopls' diagnostics to the Assistant so the cards are explained.
 *
 * Events are not always delivered in the order they were sent (the development server writes
 * each one from its own goroutine): the crash line "Segmentation fault" could arrive after
 * `run:finished`, the run was explained without it and its card never showed. Output that comes
 * after the end explains the run again, with the same exit code.
 */
export const connectAssistant = (bridge: Bridge): Unsubscribe => {
  let timer: ReturnType<typeof setTimeout> | undefined
  let late: ReturnType<typeof setTimeout> | undefined
  let finishedWith: number | null = null
  const quietly = (work: Promise<void>): void => void work.catch(() => undefined)
  const offs = [
    bridge.on('run:started', () => {
      runNumber += 1
      finishedWith = null
      clearTimeout(late)
      runDiagnostics.set([])
    }),
    bridge.on('run:finished', ({ exitCode }) => {
      finishedWith = exitCode
      quietly(explainRun(bridge, exitCode))
    }),
    bridge.on('run:output', () => {
      if (finishedWith === null) return
      const exitCode = finishedWith
      clearTimeout(late)
      late = setTimeout(() => quietly(explainRun(bridge, exitCode)), LATE_OUTPUT_DELAY_MS)
    }),
    bridge.on('lsp:diagnostics', () => {
      clearTimeout(timer)
      timer = setTimeout(() => quietly(refreshExplanations(bridge)), REFRESH_DELAY_MS)
    })
  ]
  return () => {
    clearTimeout(timer)
    clearTimeout(late)
    offs.forEach((off) => off())
  }
}
