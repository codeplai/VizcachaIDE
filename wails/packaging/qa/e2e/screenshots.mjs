// Documentation screenshots of the Wails app with the real backend:
// docs/images/wails/{editor-run,assistant,debugger}.{en,es}.png (optimized, < 1.5 MB in total).
//   node screenshots.mjs
// Each example lives in its own folder so gopls does not report "main redeclared" noise.
import fs from 'node:fs'
import path from 'node:path'
import sharp from 'sharp'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'

const OUT = path.join(L.REPO, 'docs', 'images', 'wails')
const SHOTS = path.join(L.WQ, 'shots')

const prepare = () => {
  fs.rmSync(SHOTS, { recursive: true, force: true })
  for (const [folder, file] of [['hello', 'hello/hello.go'], ['unused', 'errors/E-UNUSED-VAR/E-UNUSED-VAR.go'], ['functions', 'functions/functions.go']]) {
    fs.mkdirSync(path.join(SHOTS, folder), { recursive: true })
    fs.copyFileSync(path.join(L.REPO, 'examples', file), path.join(SHOTS, folder, path.basename(file)))
  }
}

const press = async (page, selector, text) => {
  const handle = await page.evaluateHandle(
    (sel, txt) => [...document.querySelectorAll(sel)].find((el) => el.offsetParent !== null && el.textContent.trim().startsWith(txt)),
    selector,
    text
  )
  await handle.asElement().click()
}

const openFolder = async (page, url, folder) => {
  await L.go(page, 'FilesService', 'ListTree', path.join(SHOTS, folder))
  await U.reload(page, url)
}

const optimize = async (from, to) => {
  await sharp(from).png({ palette: true, quality: 80, compressionLevel: 9, effort: 10 }).toFile(to)
}

const takeLanguage = async (lang) => {
  const T = TEXTS[lang]
  L.writeSettings({ language: lang, theme: 'light', lastFolder: path.join(SHOTS, 'hello') })
  const dev = await L.startDev()
  const edge = await L.launchEdge()
  const raw = {}
  try {
    const page = await edge.browser.newPage()
    await U.reload(page, dev.url)
    await U.openByName(page, 'hello.go')
    await U.run(page)
    await U.waitOutput(page, T.finished)
    raw['editor-run'] = await L.shot(page, `shots-${lang}-editor-run`)

    await openFolder(page, dev.url, 'unused')
    await U.openByName(page, 'E-UNUSED-VAR.go')
    await page.waitForSelector('.cm-lintRange-error', { timeout: 30000 })
    await U.run(page)
    await L.waitText(page, T.compileFailed)
    await page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout: 20000 }, T.tryThis)
    raw.assistant = await L.shot(page, `shots-${lang}-assistant`)

    await openFolder(page, dev.url, 'functions')
    await U.openByName(page, 'functions.go')
    const gutter = await page.evaluate(() => {
      const view = document.querySelector('.cm-content').cmTile.view
      const c = view.coordsAtPos(view.state.doc.line(13).from)
      const g = document.querySelector('.cm-bpgutter').getBoundingClientRect()
      return { x: g.left + g.width / 2, y: (c.top + c.bottom) / 2 }
    })
    await page.mouse.click(gutter.x, gutter.y)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 90000 }, `${T.paused} 13`)
    await page.waitForSelector('.guide .var', { timeout: 20000 })
    await L.sleep(600)
    raw.debugger = await L.shot(page, `shots-${lang}-debugger`)
  } finally {
    await edge.stop()
    dev.stop()
    await L.sleep(1500)
    L.killLeftovers()
  }
  return raw
}

const main = async () => {
  L.killLeftovers()
  L.backupSettings()
  prepare()
  fs.mkdirSync(OUT, { recursive: true })
  try {
    for (const lang of ['en', 'es']) {
      const raw = await takeLanguage(lang)
      for (const [name, file] of Object.entries(raw)) await optimize(file, path.join(OUT, `${name}.${lang}.png`))
    }
  } finally {
    L.killLeftovers()
    L.restoreSettings()
  }
  const total = fs.readdirSync(OUT).reduce((sum, f) => sum + fs.statSync(path.join(OUT, f)).size, 0)
  console.log(`${fs.readdirSync(OUT).length} images, ${(total / 1024).toFixed(0)} KB in ${OUT}`)
}

await main()
