/**
 * What the learner types goes to the shell in order. Each write is an asynchronous call that the
 * backend may serve in parallel, so sending every keystroke on its own reordered fast typing and
 * pastes ("clang++" arrived as "lcan+g+"): one write is in flight at a time, and what is typed
 * meanwhile is sent together right after it.
 */
export const orderedInput = (send: (data: string) => Promise<void>): ((data: string) => void) => {
  let pending = ''
  let sending = false
  const flush = async (): Promise<void> => {
    sending = true
    while (pending !== '') {
      const data = pending
      pending = ''
      await send(data).catch(() => undefined) // a dead session says so with terminal:exit
    }
    sending = false
  }
  return (data: string): void => {
    pending += data
    if (!sending) void flush()
  }
}
