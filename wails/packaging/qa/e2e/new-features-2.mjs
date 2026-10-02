// Second smoke test of the new features against the real Go backend (`wails dev`):
// go vet warnings, the Calls view in a real debug session, graceful Stop, Open recent,
// creating go.mod from the Go modules dialog and the delete confirmation in the Files panel.
//   node new-features-2.mjs   → screenshots and results in %TEMP%/wq
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'

const results = []
let currentPage = null
const step = async (id, run) => {
  try {
    results.push({ id, ok: true, detail: String(await run()) })
  } catch (error) {
    results.push({ id, ok: false, detail: String(error?.message ?? error) })
    try {
      await L.shot(currentPage, `nf2-fail-${id}`)
    } catch {
      /* the page may not answer */
    }
  }
  const last = results[results.length - 1]
  console.log(`${last.ok ? 'PASS' : 'FAIL'} ${id}: ${last.detail.slice(0, 240)}`)
}
const must = (condition, message) => {
  if (!condition) throw new Error(message)
}
const gutterOf = (page, line) =>
  page.evaluate((n) => {
    const view = document.querySelector('.cm-content').cmTile.view
    const c = view.coordsAtPos(view.state.doc.line(n).from)
    const gutter = document.querySelector('.cm-bpgutter').getBoundingClientRect()
    return { x: gutter.left + gutter.width / 2, y: (c.top + c.bottom) / 2 }
  }, line)
const menuItems = (page) =>
  page.$$eval('[role=menuitem]', (els) => els.map((e) => e.textContent.trim().replace(/\s+/g, ' ')))

const GRACEFUL = `package main

import (
\t"fmt"
\t"os"
\t"os/signal"
\t"time"
)

func main() {
\tsignals := make(chan os.Signal, 1)
\tsignal.Notify(signals, os.Interrupt)
\tfmt.Println("esperando la señal")
\tgo func() {
\t\t<-signals
\t\tfmt.Println("recibí la señal, limpiando")
\t\tos.Exit(0)
\t}()
\tfor {
\t\ttime.Sleep(100 * time.Millisecond)
\t}
}
`

L.killLeftovers()
L.backupSettings()
const project = L.prepareProject()
fs.mkdirSync(path.join(project, 'qa', 'graceful'), { recursive: true })
fs.writeFileSync(path.join(project, 'qa', 'graceful', 'graceful.go'), GRACEFUL)
fs.mkdirSync(path.join(project, 'qa', 'nomod'), { recursive: true })
fs.writeFileSync(path.join(project, 'qa', 'nomod', 'main.go'), 'package main\n\nfunc main() {}\n')
L.writeSettings({ lastFolder: project, firstRun: false, language: 'es' })
const dev = await L.startDev()
const edge = await L.launchEdge()
try {
  const page = await L.openApp(edge.browser, dev.url)
  currentPage = page
  await L.sleep(2000)

  await step('vet-warning', async () => {
    await U.openByName(page, 'V-PRINTF-ARGS.go')
    await U.run(page)
    await U.waitOutput(page, 'Terminó bien')
    await page.waitForFunction(() => /formato/i.test(document.querySelector('.guide')?.innerText ?? ''), { timeout: 60000 })
    const text = await page.$eval('.guide', (e) => e.innerText.replace(/\n+/g, ' | '))
    await L.shot(page, 'nf2-01-vet')
    return text.slice(0, 200)
  })

  await step('graceful-stop', async () => {
    await U.openByName(page, 'graceful.go')
    await U.run(page)
    await U.waitOutput(page, 'esperando la señal')
    await page.keyboard.down('Shift')
    await page.keyboard.press('F5')
    await page.keyboard.up('Shift')
    await U.waitOutput(page, 'Detenido', 20000)
    const output = await U.outputText(page)
    await L.shot(page, 'nf2-02-graceful-stop')
    must(output.includes('recibí la señal, limpiando'), `the program did not get the interrupt: ${output}`)
    return output.split('\n').slice(-3).join(' / ')
  })

  await step('open-recent', async () => {
    await L.clickText(page, 'button', 'Archivo')
    await page.waitForSelector('[role=menuitem]', { timeout: 5000 })
    await L.clickText(page, '[role=menuitem]', 'Abrir reciente')
    await L.sleep(600)
    const items = await menuItems(page)
    await L.shot(page, 'nf2-03-recent')
    await page.keyboard.press('Escape')
    await page.keyboard.press('Escape')
    must(items.some((t) => t.includes('graceful.go')) && items.some((t) => t.includes('V-PRINTF-ARGS.go')), items.join(' | '))
    return items.join(' | ')
  })

  await step('delete-confirm', async () => {
    const row = await page.evaluateHandle(() => [...document.querySelectorAll('.tree li, .tree button, .tree [role=treeitem]')].find((e) => e.textContent.trim() === 'leer.go'))
    const el = row.asElement()
    must(el, 'leer.go row not found')
    await el.click({ button: 'right' })
    await page.waitForSelector('[role=menuitem]', { timeout: 5000 })
    const items = await menuItems(page)
    await L.clickText(page, '[role=menuitem]', 'Eliminar')
    await page.waitForSelector('[role=dialog], [role=alertdialog]', { timeout: 5000 })
    const dialog = await page.$eval('[role=dialog], [role=alertdialog]', (e) => e.innerText.replace(/\n+/g, ' | '))
    await L.shot(page, 'nf2-05-delete-confirm')
    await L.clickText(page, '[role=dialog] button, [role=alertdialog] button', 'Cancelar')
    await L.sleep(500)
    must(fs.existsSync(U.projectFile('leer', 'leer.go')), 'Cancel deleted the file')
    return `menu: ${items.join(' | ')} — dialog: ${dialog}`
  })

  await step('modules-create', async () => {
    await L.go(page, 'FilesService', 'ListTree', path.join(project, 'qa', 'nomod'))
    await U.reload(page, dev.url, 'es')
    await page.click('.more-trigger')
    await page.waitForSelector('[role=menuitem]', { timeout: 5000 })
    await L.clickText(page, '[role=menuitem]', 'Módulos de Go')
    await page.waitForSelector('[role=dialog]', { timeout: 5000 })
    await L.clickText(page, '[role=dialog] button', 'Crear go.mod')
    await L.sleep(400)
    const input = await page.$('[role=dialog] input')
    if (input) {
      await input.click({ clickCount: 3 })
      await input.type('example.com/nomod')
      await page.keyboard.press('Enter')
    }
    const goMod = path.join(project, 'qa', 'nomod', 'go.mod')
    for (let i = 0; i < 60 && !fs.existsSync(goMod); i++) await L.sleep(500)
    await L.sleep(800)
    await L.shot(page, 'nf2-06-modules')
    must(fs.existsSync(goMod), 'go.mod was not created')
    return fs.readFileSync(goMod, 'utf8').split('\n')[0]
  })
  await step('calls-view', async () => {
    await L.go(page, 'FilesService', 'ListTree', project)
    await U.reload(page, dev.url, 'es')
    await U.openByName(page, 'functions.go')
    const where = await gutterOf(page, 25)
    await page.mouse.click(where.x, where.y)
    await L.sleep(400)
    await L.clickText(page, '.titlebar .btn', 'Depurar')
    await page.waitForFunction(() => /línea 25/.test(document.querySelector('.guide-h .chip')?.textContent ?? ''), { timeout: 90000 })
    await L.clickText(page, '[role=tab]', 'Llamadas')
    await page.waitForFunction(() => (document.querySelector('.guide')?.innerText.match(/factorial\(n=/g) ?? []).length >= 5, { timeout: 20000 })
    const titles = await page.evaluate(() => document.querySelector('.guide').innerText.match(/[a-zA-Z]+\([^)]*\)/g))
    await L.shot(page, 'nf2-04-calls')
    await page.keyboard.down('Shift')
    await page.keyboard.press('F5')
    await page.keyboard.up('Shift')
    await L.sleep(1500)
    return titles.join(' > ')
  })

} finally {
  await edge.stop()
  dev.stop()
  L.restoreSettings()
}
fs.writeFileSync(`${L.WQ}/new-features-2.json`, JSON.stringify(results, null, 2))
console.log(`\n${results.filter((r) => r.ok).length}/${results.length} steps passed`)
process.exit(results.every((r) => r.ok) ? 0 : 1)
