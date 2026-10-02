// Smoke test of the features added after the QA, against the real Go backend (`wails dev`):
// File menu, Files panel operations, external changes, console (yaegi), Calls view, ANSI output.
//   node new-features.mjs     → screenshots and results in %TEMP%/wq
import fs from 'node:fs'
import * as L from './lib.mjs'
import * as U from './ui.mjs'

const results = []
const step = async (id, run) => {
  try {
    results.push({ id, ok: true, detail: String(await run()) })
  } catch (error) {
    results.push({ id, ok: false, detail: String(error?.message ?? error) })
  }
  const last = results[results.length - 1]
  console.log(`${last.ok ? 'PASS' : 'FAIL'} ${id}: ${last.detail.slice(0, 220)}`)
}

const must = (condition, message) => {
  if (!condition) throw new Error(message)
}

L.killLeftovers()
L.backupSettings()
const project = L.prepareProject()
L.writeSettings({ lastFolder: project, firstRun: false, language: 'es' })
const dev = await L.startDev()
const edge = await L.launchEdge()
try {
  const page = await L.openApp(edge.browser, dev.url)
  await L.sleep(1500)

  await step('file-menu', async () => {
    await L.clickText(page, 'button', 'Archivo')
    await page.waitForSelector('[role=menuitem]', { timeout: 5000 })
    const items = await page.$$eval('[role=menuitem]', (els) => els.map((e) => e.textContent.trim().replace(/\s+/g, ' ')))
    await L.shot(page, 'nf-01-file-menu')
    await page.keyboard.press('Escape')
    must(items.some((t) => t.startsWith('Nuevo archivo')) && items.some((t) => t.startsWith('Guardar como')), items.join(' | '))
    return items.join(' | ')
  })

  await step('new-file-runs', async () => {
    await page.keyboard.down('Control')
    await page.keyboard.press('n')
    await page.keyboard.up('Control')
    await L.sleep(800)
    await U.setEditorText(page, 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("\\x1b[32mverde\\x1b[0m listo")\n}\n')
    await U.run(page)
    await U.waitOutput(page, 'Terminó bien')
    const colored = await page.evaluate(() => [...document.querySelectorAll('[role=log] span')].some((s) => s.textContent.includes('verde') && s.getAttribute('style')))
    await L.shot(page, 'nf-02-new-file-ansi')
    must(colored, 'ANSI color not rendered')
    return 'untitled file ran; ANSI green rendered'
  })

  await step('files-panel-ops', async () => {
    const root = project
    const created = `${root}\\nuevo_qa.go`
    await L.go(page, 'FilesService', 'CreateFile', created, 'package main\n\nfunc main() {}\n')
    must(fs.existsSync(created), 'CreateFile did not create')
    const renamed = `${root}\\renombrado_qa.go`
    await L.go(page, 'FilesService', 'Rename', created, renamed)
    must(fs.existsSync(renamed) && !fs.existsSync(created), 'Rename failed')
    await L.go(page, 'FilesService', 'MoveToTrash', renamed)
    must(!fs.existsSync(renamed), 'MoveToTrash left the file')
    let dupError = ''
    try {
      await L.go(page, 'FilesService', 'CreateFile', `${root}\\hello\\hello.go`, 'x')
    } catch (error) {
      dupError = String(error)
    }
    must(dupError, 'creating over an existing file did not fail')
    await page.click('.side .tree')
    await L.shot(page, 'nf-03-files-panel')
    return `create, rename, trash OK; duplicate refused (${dupError.slice(0, 80)})`
  })

  await step('external-change', async () => {
    await U.openByName(page, 'hello.go')
    await L.sleep(1000)
    const file = U.projectFile('hello', 'hello.go')
    const text = fs.readFileSync(file, 'utf8').replace('package main', 'package main\n\n// cambiado fuera del IDE')
    fs.writeFileSync(file, text)
    await page.waitForFunction(() => document.querySelector('.cm-content')?.innerText.includes('cambiado fuera del IDE'), { timeout: 15000 })
    return 'clean tab reloaded silently'
  })

  await step('console', async () => {
    const first = await L.go(page, 'ConsoleService', 'Eval', 'x := 20')
    const second = await L.go(page, 'ConsoleService', 'Eval', 'x * 2 + 2')
    const upper = await L.go(page, 'ConsoleService', 'Eval', 'strings.ToUpper("vizcacha")')
    must(second.result === '42', JSON.stringify(second))
    await L.clickText(page, '[role=tab]', 'Consola')
    await L.sleep(500)
    await page.type('.console textarea, .console input', '3 + 4')
    await page.keyboard.press('Enter')
    await L.sleep(1500)
    await L.shot(page, 'nf-04-console')
    return `x := 20 → ${JSON.stringify(first.result)}; x*2+2 → ${second.result}; ToUpper → ${upper.result}`
  })
} finally {
  await edge.stop()
  dev.stop()
  L.restoreSettings()
}
fs.writeFileSync(`${L.WQ}/new-features.json`, JSON.stringify(results, null, 2))
console.log(`\n${results.filter((r) => r.ok).length}/${results.length} steps passed`)
process.exit(results.every((r) => r.ok) ? 0 : 1)
