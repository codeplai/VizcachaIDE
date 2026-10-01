import type { Bridge } from '../bridge'
import type { LanguageWiring } from './extensions'
import { goToLocation } from './navigation'

/** What the editor's language extensions need: the language service and "open this location". */
export const languageWiring = (bridge: Bridge): LanguageWiring => ({
  language: bridge.language,
  openLocation: (location) => void goToLocation(bridge, location)
})
