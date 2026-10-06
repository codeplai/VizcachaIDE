import type { Bridge, Unsubscribe } from '../bridge'
import { connectAssistant } from './assistant'
import { connectCodeLanguages } from './codeLanguages'
import { connectDebug } from './debug'
import { connectDiagnostics } from './diagnostics'
import { connectUpdates } from './updates'
import { connectExternalChanges } from './externalChanges'
import { connectFrames } from './frames'
import { connectOutline } from './outline'
import { connectRun } from './run'
import { connectSettings } from './settings'
import { connectTerminal } from './terminal'

export * from './ansi'
export * from './assistant'
export * from './callArguments'
export * from './codeLanguages'
export * from './commands'
export * from './confirm'
export * from './console'
export * from './debug'
export * from './diagnostics'
export * from './fileCommands'
export * from './fileTreeCommands'
export * from './fileTreeEdit'
export * from './enabledLanguages'
export * from './files'
export * from './firstRun'
export * from './frames'
export * from './layout'
export * from './mode'
export * from './navigation'
export * from './newProject'
export * from './notice'
export * from './outline'
export * from './panelText'
export * from './output'
export * from './packages'
export * from './packageSearch'
export * from './outputLinks'
export * from './programArguments'
export * from './recentFiles'
export * from './recentOpen'
export * from './run'
export * from './runMember'
export * from './rustToolchain'
export * from './saving'
export * from './search'
export * from './searchReplace'
export * from './quickOpen'
export * from './settings'
export * from './terminal'
export * from './toolErrors'
export * from './untitled'
export * from './updates'

/** Subscribes every store to the backend events. Call the returned function to stop. */
export const connectStores = async (bridge: Bridge): Promise<Unsubscribe> => {
  const offSettings = await connectSettings(bridge)
  const offCodeLanguages = await connectCodeLanguages(bridge)
  const offUpdates = await connectUpdates(bridge)
  const offs = [
    connectRun(bridge),
    connectDebug(bridge),
    connectFrames(bridge),
    connectDiagnostics(bridge),
    connectAssistant(bridge),
    connectOutline(bridge),
    connectExternalChanges(bridge),
    connectTerminal(bridge),
    offSettings,
    offCodeLanguages,
    offUpdates
  ]
  return () => offs.forEach((off) => off())
}
export * from './workingLanguage'
