// Page-level helpers that drive the real UI (clicks, keys, waits). Selectors follow the Svelte components.
import path from 'node:path'
import * as L from './lib.mjs'

/** Loads the app (reload) and waits until the file tree is there. */
export const reload = async (page, url, lang) => {
  await page.goto(url, { waitUntil: 'load' })
  // Headless Edge (seen with 154) leaves a reloaded tab with visibilityState "hidden": no frames,
  // so requestAnimationFrame never fires and every wait or click hangs. Keep the tab in front and
  // focused.
  await page.bringToFront()
  await (await page.createCDPSession()).send('Emulation.setFocusEmulationEnabled', { enabled: true })
  await page.waitForFunction(() => !!window.go && document.querySelector('.titlebar'), { timeout: 30000 })
  await page.waitForFunction(() => document.querySelector('.crumbs b') || document.body.innerText.length > 50, { timeout: 30000 })
  await L.sleep(1500)
  if (lang) await page.evaluate(() => 0)
}

/** Double-clicks a file in the Files panel (opens it in a tab). */
export const openByName = async (page, name) => {
  const handle = await page.evaluateHandle(
    (n) => [...document.querySelectorAll('button.file:not(.dir)')].find((b) => b.textContent.trim() === n),
    name
  )
  const el = handle.asElement()
  if (!el) throw new Error(`file not in the tree: ${name}`)
  await el.evaluate((b) => b.scrollIntoView({ block: 'center' }))
  await el.click({ clickCount: 2, delay: 30 })
  await page.waitForFunction((n) => document.querySelector('.crumbs b')?.textContent === n, { timeout: 10000 }, name)
  await L.sleep(500)
}

export const run = async (page) => page.keyboard.press('F5')

export const outputText = (page) => page.evaluate(() => document.querySelector('[role=log]')?.innerText ?? '')

export const waitOutput = (page, text, timeout = 60000) =>
  page.waitForFunction((t) => (document.querySelector('[role=log]')?.innerText ?? '').includes(t), { timeout }, text)

/** Sets the editor text through the CodeMirror view (as if typed). */
export const setEditorText = (page, text) =>
  page.evaluate((t) => {
    const view = document.querySelector('.cm-content').cmTile.view
    view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: t } })
  }, text)

export const editorText = (page) => page.evaluate(() => document.querySelector('.cm-content').cmTile.view.state.doc.toString())

export const cursorLine = (page) =>
  page.evaluate(() => {
    const view = document.querySelector('.cm-content').cmTile.view
    return view.state.doc.lineAt(view.state.selection.main.head).number
  })

export const statusPosition = (page) => page.evaluate(() => document.querySelector('footer.status .sp')?.textContent.trim())

export const projectFile = (...parts) => path.join(L.PROJECT, ...parts)
