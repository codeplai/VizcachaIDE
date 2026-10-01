import type { Bridge } from '../bridge'
import { updateSettings } from './settings'
import { BLANK_PROGRAM, HELLO_PROGRAM, openUntitled } from './untitled'

export type FirstProgram = 'hello' | 'blank'

export const FIRST_RUN_STEPS = 3

/** Finishes the welcome wizard: opens the program the user picked and never shows it again. */
export const completeFirstRun = async (bridge: Bridge, program: FirstProgram): Promise<void> => {
  await openUntitled(bridge, program === 'hello' ? HELLO_PROGRAM : BLANK_PROGRAM)
  await updateSettings(bridge, { firstRun: false })
}
