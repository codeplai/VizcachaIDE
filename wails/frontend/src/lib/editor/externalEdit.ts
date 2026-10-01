import { Annotation } from '@codemirror/state'
import type { ViewUpdate } from '@codemirror/view'

/** Marks transactions that come from the app, so they are not echoed back as user edits. */
export const external = Annotation.define<boolean>()

export const isExternalEdit = (update: ViewUpdate): boolean =>
  update.transactions.some((transaction) => transaction.annotation(external))
