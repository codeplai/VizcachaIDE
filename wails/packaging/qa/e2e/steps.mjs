// The parity checks, one function per step. Each returns a short piece of evidence (a string) or throws.
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'

const must = (condition, message) => {
  if (!condition) throw new Error(message)
}

/** Clicks a visible button/menu item by its exact text. */
const press = async (page, selector, text) => {
  const handle = await page.evaluateHandle(
    (sel, txt) => [...document.querySelectorAll(sel)].find((el) => el.offsetParent !== null && el.textContent.trim().replace(/\s+/g, ' ').startsWith(txt)),
    selector,
    text
  )
  const el = handle.asElement()
  must(el, `no visible ${selector} starting with "${text}"`)
  await el.click()
}

const lineOf = (page, line) =>
  page.evaluate((n) => {
    const view = document.querySelector('.cm-content').cmTile.view
    const pos = view.state.doc.line(n).from
    const c = view.coordsAtPos(pos)
    const gutter = document.querySelector('.cm-bpgutter').getBoundingClientRect()
    return { x: gutter.left + gutter.width / 2, y: (c.top + c.bottom) / 2, textX: c.left }
  }, line)

const waitIdle = (page, ms = 400) => L.sleep(ms)

// `ctx.page` changes when the harness restarts wails dev, so steps use this forwarding proxy.
const pageOf = (ctx) =>
  new Proxy({}, { get: (_, key) => (typeof ctx.page[key] === 'function' ? ctx.page[key].bind(ctx.page) : ctx.page[key]) })

export const steps = (ctx, lang) => {
  const page = pageOf(ctx)
  const url = ctx.url
  const T = TEXTS[lang]
  const list = []
  const add = (id, title, fn) => list.push({ id, title, run: fn })
  const out = () => U.outputText(page)
  const chip = () => page.evaluate(() => document.querySelector('.guide-h .chip')?.textContent.trim() ?? '')

  add('open-folder', 'Open folder: the Files panel lists the project (ListTree on the saved folder)', async () => {
    const tree = await L.go(page, 'FilesService', 'ListTree', L.PROJECT)
    const names = tree.children.map((c) => c.name)
    must(names.includes('hello') && names.includes('errors'), `tree = ${names.join(',')}`)
    await U.reload(page, url)
    const shown = await page.$$eval('button.file', (els) => els.map((e) => e.textContent.trim()))
    must(shown.includes('hello.go') && shown.includes('functions.go'), 'Files panel does not show hello.go')
    return `${tree.children.length} entries; panel shows ${shown.length} nodes`
  })

  add('hello', 'Open hello.go, run it: output and "finished" line', async () => {
    await U.openByName(page, 'hello.go')
    await U.run(page)
    await U.waitOutput(page, T.finished)
    const text = await out()
    must(text.includes('Hello, VizcachaIDE!'), `output: ${text}`)
    await L.shot(page, `${lang}-01-hello`)
    return text.split('\n').slice(-2).join(' / ')
  })

  add('stdin', 'leer.go reads stdin: type the answers in the Output input', async () => {
    await U.openByName(page, 'leer.go')
    await U.run(page)
    await U.waitOutput(page, 'Ingresa tu nombre')
    await page.waitForSelector('input.stdin', { timeout: 5000 })
    await page.type('input.stdin', 'Ana')
    await page.keyboard.press('Enter')
    await U.waitOutput(page, 'Ingresa tu edad')
    await page.type('input.stdin', '30')
    await page.keyboard.press('Enter')
    await U.waitOutput(page, T.finished)
    const text = await out()
    must(text.includes('Hola, Ana. Tienes 30 años.'), `output: ${text}`)
    await L.shot(page, `${lang}-02-stdin`)
    return 'Hola, Ana. Tienes 30 años.'
  })

  add('args', 'Program arguments field next to Run (quotes supported)', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(L.PROJECT, 'qa', 'args'))
    await U.reload(page, url)
    await U.openByName(page, 'main.go')
    const field = await page.$('.titlebar input.args')
    must(field, 'no .titlebar input.args')
    const placeholder = await field.evaluate((e) => e.placeholder)
    must(placeholder === T.placeholderArgs, `placeholder "${placeholder}"`)
    await field.type('one "two words" three')
    await U.run(page)
    await U.waitOutput(page, T.finished)
    const text = await out()
    must(text.includes('args=[one|two words|three]'), `output: ${text}`)
    await L.shot(page, `${lang}-03-args`)
    return 'args=[one|two words|three]'
  })

  add('open-examples', 'Back to the examples folder', async () => {
    await L.go(page, 'FilesService', 'ListTree', L.PROJECT)
    await U.reload(page, url)
    return 'ok'
  })

  add('error-flow', 'E-UNUSED-VAR: exact underline, failed run, translated Assistant card, Go to line, output link', async () => {
    await U.openByName(page, 'E-UNUSED-VAR.go')
    await page.waitForSelector('.cm-lintRange-error', { timeout: 30000 })
    const underlined = await page.$$eval('.cm-lintRange-error', (els) => els.map((e) => e.textContent))
    must(underlined.join() === 'count', `underlined = ${underlined}`)
    await L.shot(page, `${lang}-04-underline`)
    await U.run(page)
    await L.waitText(page, T.compileFailed)
    await page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout: 15000 }, T.tryThis)
    const cards = await page.$$eval('article.card', (els) => els.map((e) => e.innerText))
    const card = cards.find((c) => c.includes(T.unusedTitle))
    must(card, `no translated card; cards: ${cards.map((c) => c.split('\n')[0]).join(' || ')}`)
    must(card.includes(T.goToLine), 'card has no "Go to line" button')
    await L.shot(page, `${lang}-05-assistant`)
    await page.evaluate(() => document.querySelector('.cm-content').cmTile.view.dispatch({ selection: { anchor: 0 } }))
    const button = await page.evaluateHandle(
      (title, go) => {
        const card = [...document.querySelectorAll('article.card')].find((c) => c.innerText.includes(title))
        return [...card.querySelectorAll('button')].find((b) => b.textContent.trim().startsWith(go))
      },
      T.unusedTitle,
      T.goToLine
    )
    await button.asElement().click()
    await waitIdle(page)
    const lineAfterButton = await U.cursorLine(page)
    must(lineAfterButton === 5, `Go to line moved the cursor to ${lineAfterButton}`)
    await page.evaluate(() => document.querySelector('.cm-content').cmTile.view.dispatch({ selection: { anchor: 0 } }))
    await L.clickText(page, '[role=log] button.place', 'E-UNUSED-VAR.go:5:2')
    await waitIdle(page)
    const status = await U.statusPosition(page)
    must((await U.cursorLine(page)) === 5, `output link moved the cursor to ${status}`)
    return `underline "count"; card "${card.split('\n')[0]}"; button -> line 5; link -> ${status}`
  })

  add('panic', 'P-NIL-MAP: runtime panic explained', async () => {
    await U.openByName(page, 'P-NIL-MAP.go')
    await U.run(page)
    await page.waitForFunction(() => /exit status 2/.test(document.querySelector('[role=log]')?.innerText ?? ''), { timeout: 60000 })
    await page.waitForFunction((t) => [...document.querySelectorAll('article.card h3')].some((h) => h.innerText.includes(t)), { timeout: 15000 }, T.panicTitle)
    const lines = (await out()).split('\n')
    must(lines.some((l) => l.startsWith('panic: assignment to entry in nil map')), `panic line is split: ${lines.slice(0, 4).join(' / ')}`)
    await page.evaluate(() => document.querySelector('article.card h3')?.scrollIntoView())
    await L.shot(page, `${lang}-06-panic`)
    return `card "${T.panicTitle}"; panic line in one piece; ${lines.at(-1)}`
  })

  add('debug', 'functions.go: breakpoint on 13, Debug, a=5/b=7/result=35, Next line, just changed, call stack, goroutines, Run to here, Stop debugging', async () => {
    // editor-extras leaves the Files panel on qa/complete: go back to the project first.
    await L.go(page, 'FilesService', 'ListTree', L.PROJECT)
    await U.reload(page, url)
    await U.openByName(page, 'functions.go')
    const where = await lineOf(page, 13)
    await page.mouse.click(where.x, where.y)
    await waitIdle(page)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 90000 }, `${T.paused} 13`)
    await page.waitForSelector('.guide .var', { timeout: 20000 })
    const vars = await page.$$eval('.guide .var', (els) => els.map((e) => [e.querySelector('.n')?.textContent.trim(), e.querySelector('.v')?.textContent.trim()]))
    const map = Object.fromEntries(vars)
    must(map.a === '5' && map.b === '7' && map.result === '35', `variables = ${JSON.stringify(map)}`)
    await L.shot(page, `${lang}-07-debug-paused`)
    const stack = await page.$$eval('.guide ul.stack button', (els) => els.map((e) => e.textContent.replace(/\s+/g, ' ').trim()))
    must(stack.length >= 2 && stack[0].includes('multiply'), `stack = ${stack.join(' > ')}`)
    const guideText = await page.evaluate(() => document.querySelector('.guide').innerText.toLowerCase())
    must(guideText.includes(T.callStack.toLowerCase()), 'no "How you got here" section')
    // Next line from "return result" leaves multiply and lands back in main on the call (line 36).
    await press(page, '.debugbar .db', T.nextLine)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 30000 }, `${T.paused} 36`)
    // The next one finishes the assignment: `product` appears and is marked as just changed.
    await press(page, '.debugbar .db', T.nextLine)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 30000 }, `${T.paused} 37`)
    await page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout: 20000 }, T.justChanged)
    const afterStep = await chip()
    const changed = await page.$$eval('.guide .var.changed .n', (els) => els.map((e) => e.textContent.trim()))
    must(changed.includes('product'), `just changed: ${changed}`)
    await L.shot(page, `${lang}-08-debug-step`)
    await press(page, '.guide [role=tab]', T.goroutines)
    await page.waitForSelector('.guide ul.routines button', { timeout: 10000 })
    const routines = await page.$$eval('.guide ul.routines button', (els) => els.map((e) => e.textContent.replace(/s+/g, ' ').trim()))
    await L.shot(page, `${lang}-09-debug-goroutines`)
    // Run to here: put the cursor on line 43 (fact := factorial(5)) and ask the debugger to run until there.
    await page.evaluate(() => {
      const view = document.querySelector('.cm-content').cmTile.view
      view.dispatch({ selection: { anchor: view.state.doc.line(43).from }, scrollIntoView: true })
    })
    await L.sleep(300)
    await press(page, '.debugbar .db', T.runToHere)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 30000 }, `${T.paused} 43`)
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    await page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout: 30000 }, T.run)
    return `${JSON.stringify(map)}; after two Next line: ${afterStep}, changed: ${changed}; goroutines: ${routines[0]}; run-to-here -> 43; stopped`
  })

  add('editor-intel', 'gopls: completion after "fmt.", hover and Ctrl+click to the definition', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(L.PROJECT, 'qa', 'complete'))
    await U.reload(page, url)
    await U.openByName(page, 'main.go')
    await page.waitForSelector('.cm-content', { timeout: 10000 })
    await L.sleep(2500)
    // hover on Println (line 8)
    const pos = await page.evaluate(() => {
      const view = document.querySelector('.cm-content').cmTile.view
      const line = view.state.doc.line(8)
      const at = line.from + line.text.indexOf('Println') + 3
      const c = view.coordsAtPos(at)
      return { x: c.left, y: (c.top + c.bottom) / 2 }
    })
    await page.mouse.move(pos.x, pos.y)
    await page.waitForSelector('.cm-tooltip-hover', { timeout: 20000 })
    const hover = await page.$eval('.cm-tooltip-hover', (e) => e.innerText)
    must(/Println/.test(hover), `hover = ${hover}`)
    await L.shot(page, `${lang}-10-hover`)
    await page.mouse.move(5, 5)
    // Ctrl+click on add( in line 8
    const add = await page.evaluate(() => {
      const view = document.querySelector('.cm-content').cmTile.view
      const line = view.state.doc.line(8)
      const c = view.coordsAtPos(line.from + line.text.indexOf('add') + 1)
      return { x: c.left, y: (c.top + c.bottom) / 2 }
    })
    await page.keyboard.down('Control')
    await page.mouse.click(add.x, add.y)
    await page.keyboard.up('Control')
    await waitIdle(page, 800)
    const definitionLine = await U.cursorLine(page)
    must(definitionLine === 5, `Ctrl+click went to line ${definitionLine}`)
    // completion after "fmt." on a new line
    await page.evaluate(() => {
      const view = document.querySelector('.cm-content').cmTile.view
      const line = view.state.doc.line(8)
      view.dispatch({ changes: { from: line.to, insert: '\n\tfmt.' }, selection: { anchor: line.to + 6 } })
      view.focus()
    })
    await L.sleep(600)
    await page.keyboard.down('Control')
    await page.keyboard.press('Space')
    await page.keyboard.up('Control')
    await page.waitForFunction(() => /Println/.test(document.querySelector('.cm-tooltip-autocomplete')?.innerText ?? ''), { timeout: 20000 })
    const items = await page.$eval('.cm-tooltip-autocomplete', (e) => e.innerText.split('\n').slice(0, 5).join(', '))
    await L.shot(page, `${lang}-11-completion`)
    await page.keyboard.press('Escape')
    return `hover: ${hover.split('\n')[0]}; definition line 5; completion: ${items}`
  })

  add('format-save', 'Ctrl+S formats with gofmt; closing a tab with changes asks (Save / Don\'t save / Cancel)', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(L.PROJECT, 'qa', 'fmt'))
    await U.reload(page, url)
    await U.openByName(page, 'main.go')
    const file = path.join(L.PROJECT, 'qa', 'fmt', 'main.go')
    const before = fs.readFileSync(file, 'utf8')
    await page.evaluate(() => document.querySelector('.cm-content').focus())
    await page.keyboard.down('Control')
    await page.keyboard.press('s')
    await page.keyboard.up('Control')
    await page.waitForFunction(() => true)
    await L.sleep(800)
    const after = fs.readFileSync(file, 'utf8')
    must(after !== before && after.includes('func main() {') && after.includes('\tx := 1'), `file not formatted:\n${after}`)
    const shown = await U.editorText(page)
    must(shown.replace(/\r/g, '') === after.replace(/\r/g, ''), 'editor text differs from the saved file')
    // change, then close: the dialog
    await U.setEditorText(page, shown + '// changed\n')
    await page.waitForSelector('.tab .dot', { timeout: 5000 })
    await page.click('.tab.on .close')
    await page.waitForSelector('.dlg', { timeout: 5000 })
    const question = await page.$eval('.dlg', (e) => e.innerText)
    must(question.includes(T.closeChanges), `dialog: ${question}`)
    await L.shot(page, `${lang}-12-close-dialog`)
    await press(page, '.dlg button', T.dontSave)
    await page.waitForFunction(() => !document.querySelector('.dlg'), { timeout: 5000 })
    must(fs.readFileSync(file, 'utf8') === after, "Don't save changed the file")
    const tabs = await page.$$eval('.tabs [role=tab]', (els) => els.map((e) => e.textContent.trim()))
    must(!tabs.includes('main.go'), `tab still open: ${tabs}`)
    return `formatted on disk (${before.length} -> ${after.length} bytes); dialog: "${question.split('\n').find((l) => l.includes(T.closeChanges))}"; "${T.dontSave}" kept the file`
  })

  add('stop', 'Stop a running program (qa/loop)', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(L.PROJECT, 'qa', 'loop'))
    await U.reload(page, url)
    await U.openByName(page, 'main.go')
    await U.run(page)
    await U.waitOutput(page, 'tick 3')
    await press(page, '.titlebar .btn', T.stop)
    await U.waitOutput(page, T.stopped)
    const before = (await out()).split('\n').length
    await L.sleep(800)
    must((await out()).split('\n').length === before, 'the program kept printing after Stop')
    return 'Stopped, no more output'
  })

  add('modules', 'Module run: go run . in a folder with go.mod (qa/mod)', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(L.PROJECT, 'qa', 'mod'))
    await U.reload(page, url)
    await U.openByName(page, 'main.go')
    await U.run(page)
    await U.waitOutput(page, T.finished)
    must((await out()).includes('module ok'), `output: ${await out()}`)
    return 'module ok (two files, needs go run .)'
  })

  add('editor-extras', 'Find, Go to line, toggle comment, zoom, Outline', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(L.PROJECT, 'qa', 'complete'))
    await U.reload(page, url)
    await U.openByName(page, 'main.go')
    await page.evaluate(() => document.querySelector('.cm-content').focus())
    await page.keyboard.down('Control'); await page.keyboard.press('f'); await page.keyboard.up('Control')
    await page.waitForSelector('.cm-search', { timeout: 5000 })
    await page.keyboard.press('Escape')
    await page.keyboard.down('Control'); await page.keyboard.press('g'); await page.keyboard.up('Control')
    await page.waitForSelector('.cm-goto-line', { timeout: 5000 })
    await page.keyboard.press('Escape')
    await page.evaluate(() => { const v = document.querySelector('.cm-content').cmTile.view; v.dispatch({ selection: { anchor: v.state.doc.line(8).from } }); v.focus() })
    await page.keyboard.down('Control'); await page.keyboard.press('/'); await page.keyboard.up('Control')
    const commented = (await U.editorText(page)).split('\n')[7]
    must(commented.trim().startsWith('//'), `toggle comment: ${commented}`)
    await page.keyboard.down('Control'); await page.keyboard.press('/'); await page.keyboard.up('Control')
    const sizeBefore = await page.$eval('.cm-content', (e) => parseFloat(getComputedStyle(e).fontSize))
    await page.keyboard.down('Control'); await page.keyboard.press('='); await page.keyboard.up('Control')
    await L.sleep(500)
    const sizeAfter = await page.$eval('.cm-content', (e) => parseFloat(getComputedStyle(e).fontSize))
    must(sizeAfter > sizeBefore, `zoom ${sizeBefore} -> ${sizeAfter}`)
    await page.keyboard.down('Control'); await page.keyboard.press('0'); await page.keyboard.up('Control')
    await page.click('.rail button:nth-child(2)')
    await page.waitForFunction(() => /add/.test(document.querySelector('.side')?.innerText ?? ''), { timeout: 15000 })
    const outline = await page.$eval('.side', (e) => e.innerText.replace(/\s+/g, ' '))
    await L.shot(page, `${lang}-13-outline`)
    await page.click('.rail button:nth-child(1)')
    return `find, goto, comment toggled, zoom ${sizeBefore}->${sizeAfter}px, outline: ${outline}`
  })

  add('layout-1024', 'At the narrowest window (1024 px) the title bar and the debug toolbar do not overflow', async () => {
    await L.go(page, 'FilesService', 'ListTree', L.PROJECT)
    await U.reload(page, url)
    await U.openByName(page, 'functions.go')
    const where = await lineOf(page, 13)
    await page.mouse.click(where.x, where.y)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 90000 }, `${T.paused} 13`)
    // The narrowest supported window (1024 px): nothing may overflow or overlap while debugging.
    await page.setViewport({ width: 1024, height: 640 })
    await L.sleep(600)
    const layout = await page.evaluate(() => {
      const box = (selector) => document.querySelector(selector).getBoundingClientRect()
      const title = document.querySelector('.titlebar')
      return {
        pageOverflow: document.documentElement.scrollWidth - window.innerWidth,
        titleOverflow: title.scrollWidth - title.clientWidth,
        toolbarInsideEditor: box('.debugbar').left >= box('.editor').left - 1 && box('.debugbar').right <= box('.editor').right + 1,
        toolbarWidth: Math.round(box('.debugbar').width),
        editorWidth: Math.round(box('.editor').width)
      }
    })
    await L.shot(page, `${lang}-14-width-1024`)
    await page.setViewport({ width: 1280, height: 800 })
    must(layout.pageOverflow <= 0 && layout.titleOverflow <= 0, `overflow at 1024 px: ${JSON.stringify(layout)}`)
    must(layout.toolbarInsideEditor, `debug toolbar leaves the editor at 1024 px: ${JSON.stringify(layout)}`)
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    return `no overflow; toolbar ${layout.toolbarWidth}px inside an editor of ${layout.editorWidth}px`
  })

  // The debugger steps go last: after them the harness' page sometimes stops answering (see QA_WAILS.md).
  // The desktop window of `wails dev` sometimes loses its WebView2 under load, and the events stop
  // reaching the browser tab (QA_WAILS.md): the harness restarts wails dev between groups of steps.
  const restart = (id) => ({ id: `restart-${id}`, silent: true, run: () => ctx.restart() })
  const ordered = [...list.filter((s) => !['debug', 'layout-1024'].includes(s.id)), ...list.filter((s) => ['debug', 'layout-1024'].includes(s.id))]
  const index = (id) => ordered.findIndex((s) => s.id === id)
  ordered.splice(index('modules'), 0, restart('modules'))
  ordered.splice(index('editor-intel'), 0, restart('intel'))
  return ordered
}
