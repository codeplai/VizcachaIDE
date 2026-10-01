// Public surface of the editor for the shell and the panels.
export { default as EditorPane } from './EditorPane.svelte'
export { goToLocation, gotoRequest, type GotoRequest } from './navigation'
export { saveDocument, saveNotice, type SaveNotice } from './saving'
export {
  answerCloseRequest,
  closeRequest,
  requestCloseTab,
  saveActiveDocument,
  tabItems,
  type CloseAnswer,
  type CloseRequest,
  type TabItem
} from './tabs'
export { editorFontSize, zoomEditor, type ZoomAction } from './zoom'
