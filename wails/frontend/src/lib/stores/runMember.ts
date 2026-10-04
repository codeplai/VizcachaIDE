// Choosing which program of a Cargo workspace to run: the backend answers a run with an error
// "run.chooseMember: app, cli" and waits for the member the student picks.
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { RunApi } from '../bridge/types'
import { resetRun } from './run'
import { withToolErrors } from './toolErrors'

/** The key the backend's error message starts with (the i18n key of the dialog's title). */
export const CHOOSE_MEMBER_KEY = 'run.chooseMember'

/** The bridge call that runs the chosen member. */
export type RunMemberApi = Pick<RunApi, 'runMember'>

/** The run waiting for a member: what was asked and the members to choose from. */
export interface MemberChoice {
  path: string
  args: string[]
  members: string[]
}

export const memberChoice = writable<MemberChoice | null>(null)

/** The member names of a "run.chooseMember: app, cli" message; null when it is another error. */
export const membersIn = (error: unknown): string[] | null => {
  const message = error instanceof Error ? error.message : String(error ?? '')
  const at = message.indexOf(CHOOSE_MEMBER_KEY)
  if (at < 0) return null
  const names = message
    .slice(at + CHOOSE_MEMBER_KEY.length)
    .replace(/^\s*:/, '')
    .split(',')
    .map((name) => name.trim())
    .filter(Boolean)
  return names.length > 0 ? names : null
}

/** Runs `call`; when the backend asks for a workspace member, opens the dialog instead of failing. */
export const runOrChooseMember = async (
  path: string,
  args: string[],
  call: () => Promise<unknown>
): Promise<void> => {
  try {
    await call()
  } catch (error) {
    const members = membersIn(error)
    if (!members) throw error
    memberChoice.set({ path, args, members })
  }
}

/** Runs the chosen member and closes the dialog. */
export const chooseMember = async (bridge: Bridge, member: string): Promise<void> => {
  const choice = get(memberChoice)
  memberChoice.set(null)
  if (!choice) return
  const { path, args } = choice
  resetRun()
  await withToolErrors(bridge, 'rust', () => bridge.run.runMember(path, member, args))
}

export const dismissMemberChoice = (): void => memberChoice.set(null)
