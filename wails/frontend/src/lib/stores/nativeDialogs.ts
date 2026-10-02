// The native Open / Save as / folder dialogs run in the Go backend. If Windows refuses to open
// one, the user must see why instead of a button that seems to do nothing.
import { showNotice } from './notice'

const reasonOf = (error: unknown): string =>
  error instanceof Error ? error.message : String(error ?? '')

/** Runs a dialog call; on failure shows a notice and returns `fallback` (as if cancelled). */
export const askNativeDialog = async <T>(call: () => Promise<T>, fallback: T): Promise<T> => {
  try {
    return await call()
  } catch (error) {
    showNotice({
      messageKey: 'files.dialogFailed',
      values: { reason: reasonOf(error) },
      actions: []
    })
    return fallback
  }
}
