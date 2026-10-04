import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { startingCodeLanguage } from './enabledLanguages'
import { settings, updateSettings } from './settings'
import { openUntitled } from './untitled'

export type FirstProgram = 'hello' | 'blank'

export const FIRST_RUN_STEPS = 4

/**
 * Finishes the welcome wizard: opens the program the user picked, in the first language they
 * chose, and never shows the wizard again.
 */
export const completeFirstRun = async (bridge: Bridge, program: FirstProgram): Promise<void> => {
  await openUntitled(bridge, startingCodeLanguage(get(settings)?.defaultCodeLanguage), program)
  await updateSettings(bridge, { firstRun: false })
}
