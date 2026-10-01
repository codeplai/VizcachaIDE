import type { Bridge, Unsubscribe } from '../bridge'
import { connectDebug } from './debug'
import { connectDiagnostics } from './diagnostics'
import { connectRun } from './run'
import { connectSettings } from './settings'

export * from './commands'
export * from './debug'
export * from './diagnostics'
export * from './files'
export * from './layout'
export * from './mode'
export * from './outline'
export * from './output'
export * from './run'
export * from './settings'

/** Subscribes every store to the backend events. Call the returned function to stop. */
export const connectStores = async (bridge: Bridge): Promise<Unsubscribe> => {
  const offSettings = await connectSettings(bridge)
  const offs = [connectRun(bridge), connectDebug(bridge), connectDiagnostics(bridge), offSettings]
  return () => offs.forEach((off) => off())
}
