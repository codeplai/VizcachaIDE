/// <reference types="vite/client" />

/** VizcachaIDE version, from wails.json (set by vite.config.ts). */
declare const __APP_VERSION__: string

interface Window {
  /** Injected by Wails when the app runs inside the desktop window. */
  go?: unknown
}
