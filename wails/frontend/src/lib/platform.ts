// The operating system the IDE runs on, for the few hints that differ per system.
export type SystemPlatform = 'windows' | 'macos' | 'linux'

/** Reads the platform from what the webview says about itself (Wails shows the system's own). */
export const detectPlatform = (): SystemPlatform => {
  if (typeof navigator === 'undefined') return 'linux'
  const hint = `${navigator.platform ?? ''} ${navigator.userAgent ?? ''}`.toLowerCase()
  if (hint.includes('win')) return 'windows'
  if (hint.includes('mac')) return 'macos'
  return 'linux'
}
