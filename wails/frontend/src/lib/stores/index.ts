import type { Bridge, Unsubscribe } from '../bridge'
import { connectAssistant } from './assistant'
import { connectDebug } from './debug'
import { connectDiagnostics } from './diagnostics'
import { connectOutline } from './outline'
import { connectRun } from './run'
import { connectSettings } from './settings'

export * from './assistant'
export * from './commands'
export * from './confirm'
export * from './debug'
export * from './diagnostics'
export * from './files'
export * from './firstRun'
export * from './layout'
export * from './mode'
export * from './navigation'
export * from './notice'
export * from './outline'
export * from './output'
export * from './outputLinks'
export * from './programArguments'
export * from './run'
export * from './saving'
export * from './settings'
export * from './untitled'

/** Subscribes every store to the backend events. Call the returned function to stop. */
export const connectStores = async (bridge: Bridge): Promise<Unsubscribe> => {
  const offSettings = await connectSettings(bridge)
  const offs = [
    connectRun(bridge),
    connectDebug(bridge),
    connectDiagnostics(bridge),
    connectAssistant(bridge),
    connectOutline(bridge),
    offSettings
  ]
  return () => offs.forEach((off) => off())
}
