// User actions. They call the bridge and update stores; components only call these.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { Breakpoint } from '../domain'
import { breakpoints, debugActive, debugStarting } from './debug'
import { activePath, buffers } from './files'
import { openDialog } from './layout'
import { showNotice } from './notice'
import { resetRun, stoppedByUser } from './run'
import { isUntitled } from './untitled'

export const GO_DOWNLOAD_URL = 'https://go.dev/dl/'
export const DELVE_INSTALL_COMMAND = 'go install github.com/go-delve/delve/cmd/dlv@latest'

const breakpointsOf = (file: string): Breakpoint[] =>
  (get(breakpoints)[file] ?? []).map((line) => ({
    location: { file, line, column: 1 },
    condition: ''
  }))

type MissingTool = 'go' | 'delve'

/** Reads the backend's failure text and tells which tool is missing, if any. */
export const missingToolIn = (message: string): MissingTool | null => {
  const text = message.toLowerCase()
  if (/\b(dlv|delve)\b/.test(text)) return 'delve'
  return /\bgo\b.*(not found|not installed|no such file)/.test(text) ? 'go' : null
}

const reasonOf = (error: unknown): string =>
  error instanceof Error ? error.message : String(error ?? '')

const explainMissingTool = (bridge: Bridge, tool: MissingTool): void => {
  if (tool === 'delve') {
    showNotice({
      messageKey: 'errors.delveNotFound',
      values: {},
      detail: DELVE_INSTALL_COMMAND,
      actions: [
        {
          labelKey: 'errors.delveCopyCommand',
          run: () => void navigator.clipboard?.writeText(DELVE_INSTALL_COMMAND)
        }
      ]
    })
    return
  }
  showNotice({
    messageKey: 'errors.goNotFound',
    values: {},
    actions: [
      { labelKey: 'errors.goNotFoundInstall', run: () => bridge.system.openUrl(GO_DOWNLOAD_URL) },
      { labelKey: 'errors.goNotFoundChoose', run: () => openDialog.set('settings') }
    ]
  })
}

/** Runs `action`; when the backend says a tool is missing, shows the matching IDE message. */
const withToolErrors = async (bridge: Bridge, action: () => Promise<unknown>): Promise<void> => {
  try {
    await action()
  } catch (error) {
    debugStarting.set(false)
    const tool = missingToolIn(reasonOf(error))
    if (!tool) throw error
    explainMissingTool(bridge, tool)
  }
}

export const runActiveFile = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (!path) return
  resetRun()
  await withToolErrors(bridge, () =>
    isUntitled(path)
      ? bridge.run.runUntitled(get(buffers)[path] ?? '', [])
      : bridge.run.run(path, [])
  )
}

export const stopProgram = (bridge: Bridge): Promise<void> => {
  stoppedByUser.set(true)
  return bridge.run.stop()
}

export const startDebugging = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (!path || get(debugActive) || isUntitled(path)) return
  debugStarting.set(true)
  await withToolErrors(bridge, () => bridge.debug.start(path, breakpointsOf(path)))
}

export const stopDebugging = (bridge: Bridge): Promise<void> => bridge.debug.stop()
export const stepOver = (bridge: Bridge): Promise<void> => bridge.debug.stepOver()
export const stepInto = (bridge: Bridge): Promise<void> => bridge.debug.stepInto()
export const stepOut = (bridge: Bridge): Promise<void> => bridge.debug.stepOut()
export const resumeDebugging = (bridge: Bridge): Promise<void> => bridge.debug.resume()
