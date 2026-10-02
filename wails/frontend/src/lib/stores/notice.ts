import { writable } from 'svelte/store'

export interface NoticeAction {
  labelKey: string
  run: () => void
}

/** A message about the IDE itself (not about the user's code) with what to do next. */
export interface Notice {
  messageKey: string
  values: Record<string, string | number>
  actions: NoticeAction[]
  /** Monospace text under the message, for example a command to copy. */
  detail?: string
  /** `info` is a calm confirmation (for example "Path copied."); the default is a problem. */
  tone?: 'info'
}

export const notice = writable<Notice | null>(null)

export const showNotice = (next: Notice): void => notice.set(next)
export const dismissNotice = (): void => notice.set(null)
