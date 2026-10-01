// Per-viewer conveniences kept in the browser storage. The app works without it.
const PREFIX = 'vizcacha.'

export const readPreference = <T>(key: string, fallback: T): T => {
  try {
    const raw = localStorage.getItem(PREFIX + key)
    return raw === null ? fallback : (JSON.parse(raw) as T)
  } catch {
    return fallback
  }
}

export const writePreference = (key: string, value: unknown): void => {
  try {
    localStorage.setItem(PREFIX + key, JSON.stringify(value))
  } catch {
    // Private windows or blocked storage: the preference lasts until the app closes.
  }
}
