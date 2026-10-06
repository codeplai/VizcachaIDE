import { writable } from 'svelte/store'

export type ChoiceTone = 'primary' | 'plain' | 'danger'

export interface Choice<Id extends string = string> {
  id: Id
  labelKey: string
  values?: Record<string, string | number>
  tone: ChoiceTone
}

export interface ConfirmRequest {
  messageKey: string
  values: Record<string, string | number>
  choices: Choice[]
  /** Id answered when the dialog is dismissed with Escape. */
  dismissId: string
  answer: (id: string) => void
}

export const pendingConfirm = writable<ConfirmRequest | null>(null)

const ask = <Id extends string>(
  messageKey: string,
  values: Record<string, string | number>,
  choices: Choice<Id>[],
  dismissId: Id
): Promise<Id> =>
  new Promise((resolve) => {
    const answer = (id: string): void => {
      pendingConfirm.set(null)
      resolve(id as Id)
    }
    pendingConfirm.set({ messageKey, values, choices, dismissId, answer })
  })

export const confirmCloseChanges = (file: string) =>
  ask(
    'confirm.closeChanges',
    { file },
    [
      { id: 'save', labelKey: 'confirm.save', tone: 'primary' },
      { id: 'dontSave', labelKey: 'confirm.dontSave', tone: 'plain' },
      { id: 'cancel', labelKey: 'confirm.cancel', tone: 'plain' }
    ],
    'cancel'
  )

export const confirmReloadFile = (file: string) =>
  ask(
    'confirm.reloadFile',
    { file },
    [
      { id: 'reload', labelKey: 'confirm.reload', tone: 'danger' },
      { id: 'keep', labelKey: 'confirm.keepMine', tone: 'primary' }
    ],
    'keep'
  )

export const confirmReplaceAll = (count: number, file: string) =>
  ask(
    'confirm.replaceAll',
    { count, file },
    [
      { id: 'replace', labelKey: 'confirm.replaceCount', values: { count }, tone: 'primary' },
      { id: 'cancel', labelKey: 'confirm.cancel', tone: 'plain' }
    ],
    'cancel'
  )

/** Asks before "Replace all" in the Search panel; it names how many matches in how many files. */
export const confirmReplaceInFiles = (count: number, files: number) =>
  ask(
    'confirm.replaceInFiles',
    { count, files },
    [
      { id: 'replace', labelKey: 'confirm.replaceCount', values: { count }, tone: 'primary' },
      { id: 'cancel', labelKey: 'confirm.cancel', tone: 'plain' }
    ],
    'cancel'
  )

export interface DeleteDetails {
  /** Files inside the folder being deleted (undefined for a single file). */
  files?: number
  /** Some of the affected files have changes that were not saved. */
  unsaved?: boolean
}

const deleteMessageKey = (details: DeleteDetails): string => {
  if (details.unsaved) return 'confirm.deleteUnsaved'
  return details.files === undefined ? 'confirm.delete' : 'confirm.deleteFolder'
}

/** Asks before sending a file or folder to the Recycle Bin. */
export const confirmDelete = (name: string, details: DeleteDetails = {}) =>
  ask(
    deleteMessageKey(details),
    { name, count: details.files ?? 0 },
    [
      { id: 'delete', labelKey: 'confirm.moveToTrash', tone: 'danger' },
      { id: 'cancel', labelKey: 'confirm.cancel', tone: 'plain' }
    ],
    'cancel'
  )
