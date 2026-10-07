import type { Bridge } from '../bridge'
import { goToLocation } from '../stores/navigation'
import { renameSymbol, prepareRename } from '../stores/refactor'
import { hasFreshRename, hasFreshUndo, redoRename, undoRename } from '../stores/refactorUndo'
import { findReferences } from '../stores/references'
import type { LanguageWiring } from './extensions'

/** What the editor's language extensions need: the language service and "open this location". */
export const languageWiring = (bridge: Bridge): LanguageWiring => ({
  language: bridge.language,
  openLocation: (location) => void goToLocation(bridge, location),
  refactor: {
    prepareRename: (at) => prepareRename(bridge, at),
    renameSymbol: (at, newName) => renameSymbol(bridge, at, newName),
    findReferences: (at, symbol) => findReferences(bridge, at, symbol),
    undoRename: () => {
      if (!hasFreshRename()) return false
      void undoRename(bridge)
      return true
    },
    redoRename: () => {
      if (!hasFreshUndo()) return false
      void redoRename(bridge)
      return true
    }
  }
})
