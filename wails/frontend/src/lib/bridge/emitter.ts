import type { EventName, EventPayloads } from '../events'
import type { Unsubscribe } from './types'

type Handler<E extends EventName> = (payload: EventPayloads[E]) => void

/** A minimal typed event hub, used by the mock bridge in place of the Wails runtime. */
export const createEmitter = () => {
  const handlers = new Map<EventName, Set<Handler<EventName>>>()

  const on = <E extends EventName>(name: E, handler: Handler<E>): Unsubscribe => {
    const set = handlers.get(name) ?? new Set()
    set.add(handler as Handler<EventName>)
    handlers.set(name, set)
    return () => {
      set.delete(handler as Handler<EventName>)
    }
  }

  const emit = <E extends EventName>(name: E, payload: EventPayloads[E]): void => {
    handlers.get(name)?.forEach((handler) => handler(payload))
  }

  return { on, emit }
}
