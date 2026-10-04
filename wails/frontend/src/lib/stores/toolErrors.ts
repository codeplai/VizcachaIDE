// What to tell the user when the backend says a tool is missing or an action is unsupported.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { CodeLanguage, LanguageProfile, ToolSpec } from '../domain'
import { t } from '../i18n'
import { profiles } from './codeLanguages'
import { debugStarting, debuggedPath } from './debug'
import { openDialog } from './layout'
import { showNotice, type NoticeAction } from './notice'

export const GO_DOWNLOAD_URL = 'https://go.dev/dl/'
export const DELVE_INSTALL_COMMAND = 'go install github.com/go-delve/delve/cmd/dlv@latest'

/** app.MissingTool wraps the id like this: `tool "dlv": tool not found`. */
const TOOL_ID = /\btool "([^"]+)"/
const UNSUPPORTED = /does not support the action/i

/**
 * Reads the backend's failure text and tells which tool is missing, if any: the id of the ToolSpec
 * (`tool "dlv": ...`). A 2.0 backend answers with free text, which is still recognised for Go.
 */
export const missingToolIn = (message: string): string | null => {
  const id = TOOL_ID.exec(message)?.[1]
  if (id) return id
  const text = message.toLowerCase()
  if (/\b(dlv|delve)\b/.test(text)) return 'dlv'
  return /\bgo\b.*(not found|not installed|no such file)/.test(text) ? 'go' : null
}

const specOf = (id: string, all: LanguageProfile[]): ToolSpec | undefined =>
  all.flatMap((profile) => profile.tools).find((spec) => spec.id === id)

const copyAction = (command: string): NoticeAction => ({
  labelKey: 'errors.toolCopyCommand',
  run: () => void navigator.clipboard?.writeText(command)
})

const chooseAction: NoticeAction = {
  labelKey: 'errors.goNotFoundChoose',
  run: () => openDialog.set('settings')
}

const explainSpec = (bridge: Bridge, spec: ToolSpec): void => {
  const actions: NoticeAction[] = []
  if (spec.installUrl) {
    actions.push({
      labelKey: 'errors.toolInstall',
      run: () => bridge.system.openUrl(spec.installUrl)
    })
  }
  if (spec.installCommand) actions.push(copyAction(spec.installCommand))
  if (!spec.providedBy) actions.push(chooseAction)
  showNotice({
    messageKey: spec.missingKey || 'errors.toolNotFound',
    values: { tool: spec.id },
    detail: spec.installCommand || undefined,
    actions
  })
}

/** The messages of 2.0, for a backend that does not name the tool. */
const explainLegacyGo = (bridge: Bridge, id: string): void => {
  if (id === 'dlv') {
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
      chooseAction
    ]
  })
}

/** Shows the notice of a missing tool: from its ToolSpec when the profiles know it. */
export const explainMissingTool = (bridge: Bridge, id: string, namedByBackend: boolean): void => {
  const spec = specOf(id, get(profiles))
  if (spec && namedByBackend) return explainSpec(bridge, spec)
  if (id === 'go' || id === 'dlv') return explainLegacyGo(bridge, id)
  if (spec) return explainSpec(bridge, spec)
  showNotice({ messageKey: 'errors.toolNotFound', values: { tool: id }, actions: [chooseAction] })
}

const nameOf = (codeLanguage: CodeLanguage): string => {
  const profile = get(profiles).find((candidate) => candidate.id === codeLanguage)
  if (!profile) return codeLanguage
  try {
    return get(t)(profile.nameKey)
  } catch {
    return codeLanguage
  }
}

const reasonOf = (error: unknown): string =>
  error instanceof Error ? error.message : String(error ?? '')

/**
 * Runs `action` (an action on a file of `codeLanguage`); when the backend says a tool is missing or
 * the language lacks the action, shows the matching IDE message. Other failures are thrown.
 */
export const withToolErrors = async (
  bridge: Bridge,
  codeLanguage: CodeLanguage | null,
  action: () => Promise<unknown>
): Promise<void> => {
  try {
    await action()
  } catch (error) {
    debugStarting.set(false)
    debuggedPath.set(null)
    const reason = reasonOf(error)
    const tool = missingToolIn(reason)
    if (tool) return explainMissingTool(bridge, tool, TOOL_ID.test(reason))
    if (!UNSUPPORTED.test(reason)) throw error
    showNotice({
      messageKey: 'errors.unsupportedAction',
      values: { codeLanguage: codeLanguage ? nameOf(codeLanguage) : '' },
      actions: []
    })
  }
}
