// Interactions the language phases share: clicking by text, the breakpoint gutter, answering a
// program in Output and a page proxy that follows ctx.page across restarts.
import * as L from './lib.mjs'

export const must = (condition, message) => {
  if (!condition) throw new Error(message)
}

/** Clicks a visible element by the start of its text. */
export const press = async (page, selector, text) => {
  const handle = await page.evaluateHandle(
    (sel, txt) => [...document.querySelectorAll(sel)].find((el) => el.offsetParent !== null && el.textContent.trim().replace(/\s+/g, ' ').startsWith(txt)),
    selector,
    text
  )
  const el = handle.asElement()
  must(el, `no visible ${selector} starting with "${text}"`)
  await el.click()
}

/** Clicks the breakpoint gutter of a line. */
export const toggleBreakpoint = async (page, line) => {
  const where = await page.evaluate((n) => {
    const view = document.querySelector('.cm-content').cmTile.view
    const c = view.coordsAtPos(view.state.doc.line(n).from)
    const gutter = document.querySelector('.cm-bpgutter').getBoundingClientRect()
    return { x: gutter.left + gutter.width / 2, y: (c.top + c.bottom) / 2 }
  }, line)
  await page.mouse.click(where.x, where.y)
  await L.sleep(400)
}

export const answer = async (page, text) => {
  await page.waitForSelector('input.stdin:not([disabled])', { timeout: 15000 })
  await page.type('input.stdin', text)
  await page.keyboard.press('Enter')
}

export const pageOf = (ctx) =>
  new Proxy({}, { get: (_, key) => (typeof ctx.page[key] === 'function' ? ctx.page[key].bind(ctx.page) : ctx.page[key]) })
