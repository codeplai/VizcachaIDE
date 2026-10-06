// Rename symbol (F2) and Find all references (Shift+F12), against the real gopls: a small module
// with helper.go (declares Greet) and main.go (calls it twice, the first call after an emoji and
// an accent). Only main.go is open, so helper.go is edited on disk. Checks: the inline box opens
// prefilled, Enter renames in both files and says how much, Ctrl+Z undoes the open file, a keyword
// is refused with a translated message, and Shift+F12 lists the uses grouped by file and opens one.
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { must, pageOf } from './interactions.mjs'

const HELPER = 'package main\n\n// Greet says hello to a name.\nfunc Greet(name string) string {\n\treturn "hola " + name\n}\n'
const MAIN = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("😀 año", Greet("vizcacha"))\n\tfmt.Println(Greet("otra"))\n}\n'

const NOTICE = {
  en: /Renamed in 3 places across 2 files/,
  es: /Se renombró en 3 lugares de 2 archivos/
}
const REFUSED = {
  en: /cannot be renamed/,
  es: /no se puede renombrar/
}
const REFERENCES_TAB = { en: 'References', es: 'Referencias' }
const SUMMARY = { en: /3 uses of “Greet” in 2 files/, es: /3 usos de «Greet» en 2 archivos/ }

/** Mouse-free: puts the cursor inside the first occurrence of a word on a line. */
const cursorOn = (page, line, word) =>
  page.evaluate(
    (n, w) => {
      const view = document.querySelector('.cm-content').cmTile.view
      const text = view.state.doc.line(n)
      const at = text.text.indexOf(w)
      if (at < 0) return false
      view.dispatch({ selection: { anchor: text.from + at + 1 } })
      view.focus()
      return true
    },
    line,
    word
  )

const noticeText = (page) => page.evaluate(() => document.querySelector('.notice')?.innerText ?? '')

export const refactorSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const root = path.join(L.PROJECT, 'qa', 'refactor')
  const list = []
  const add = (id, title, run) => list.push({ id: `refactor-${id}`, title, run })

  add('open', 'main.go of a two-file module is open (helper.go stays closed)', async () => {
    fs.rmSync(root, { recursive: true, force: true })
    fs.mkdirSync(root, { recursive: true })
    fs.writeFileSync(path.join(root, 'go.mod'), 'module example.com/refactor\n\ngo 1.21\n')
    fs.writeFileSync(path.join(root, 'helper.go'), HELPER)
    fs.writeFileSync(path.join(root, 'main.go'), MAIN)
    await L.go(page, 'FilesService', 'ListTree', root)
    await U.reload(page, ctx.url)
    await U.openByName(page, 'main.go', 'refactor/')
    await page.waitForSelector('.cm-content', { timeout: 10000 })
    must((await U.editorText(page)).includes('Greet("otra")'), 'main.go is not on screen')
    return 'main.go open'
  })

  add('box', 'F2 opens an input at the symbol, prefilled and selected', async () => {
    let value = ''
    for (let attempt = 0; attempt < 24 && value !== 'Greet'; attempt++) {
      await L.sleep(2500) // gopls loads the module
      must(await cursorOn(page, 7, 'Greet'), 'Greet not on line 7')
      await page.keyboard.press('F2')
      value = await page
        .waitForSelector('.cm-rename input', { timeout: 4000 })
        .then((input) => input.evaluate((el) => el.value))
        .catch(() => '')
    }
    must(value === 'Greet', `the box has "${value}"`)
    const selected = await page.evaluate(() => {
      const input = document.querySelector('.cm-rename input')
      return document.activeElement === input && input.selectionStart === 0 && input.selectionEnd === input.value.length
    })
    must(selected, 'the name is not selected in the focused input')
    await L.shot(page, `refactor-${lang}-01-box`)
    await page.keyboard.press('Escape')
    await page.waitForFunction(() => !document.querySelector('.cm-rename'), { timeout: 5000 })
    return 'box opened and Escape closed it'
  })

  add('rename', 'Enter renames in main.go (buffer, unsaved) and helper.go (disk), and says how much', async () => {
    must(await cursorOn(page, 7, 'Greet'), 'Greet not on line 7')
    await page.keyboard.press('F2')
    await page.waitForSelector('.cm-rename input', { timeout: 10000 })
    await page.keyboard.type('Salute')
    await page.keyboard.press('Enter')
    await page.waitForFunction(() => document.querySelector('.cm-content').cmTile.view.state.doc.toString().includes('Salute("otra")'), { timeout: 30000 })
    const text = await U.editorText(page)
    must(text.includes('Println("😀 año", Salute("vizcacha"))') && !text.includes('Greet'), `main.go = ${text}`)
    await page.waitForFunction((re) => new RegExp(re).test(document.querySelector('.notice')?.innerText ?? ''), { timeout: 10000 }, NOTICE[lang].source)
    const helper = fs.readFileSync(path.join(root, 'helper.go'), 'utf8')
    must(helper.includes('func Salute(name string) string') && !helper.includes('Greet('), `helper.go = ${helper}`)
    const unsaved = await page.evaluate(() => document.querySelector('.tab.active, [role=tab][aria-selected=true]')?.innerText ?? '')
    await L.shot(page, `refactor-${lang}-02-renamed`)
    return `notice: ${(await noticeText(page)).split('\n')[0]}; tab: ${unsaved.trim()}`
  })

  add('undo', 'Ctrl+Z undoes the rename of the open file in one step', async () => {
    await page.click('.cm-content')
    await page.keyboard.down('Control')
    await page.keyboard.press('z')
    await page.keyboard.up('Control')
    await L.sleep(500)
    const text = await U.editorText(page)
    must(text === MAIN, `after undo main.go = ${text}`)
    return 'main.go is back to Greet'
  })

  add('keyword', 'F2 on a keyword is refused with a translated message', async () => {
    must(await cursorOn(page, 5, 'func'), 'func not on line 5')
    await page.keyboard.press('F2')
    await page.waitForFunction((re) => new RegExp(re).test(document.querySelector('.notice')?.innerText ?? ''), { timeout: 15000 }, REFUSED[lang].source)
    must(!(await page.$('.cm-rename')), 'a box opened for a keyword')
    return (await noticeText(page)).split('\n')[0]
  })

  add('references', 'Shift+F12 lists the uses grouped by file and a click opens the file there', async () => {
    // The rename above changed helper.go on disk: put it back and tell gopls via a fresh reload.
    fs.writeFileSync(path.join(root, 'helper.go'), HELPER)
    await L.go(page, 'FilesService', 'ListTree', root)
    await U.reload(page, ctx.url)
    await U.openByName(page, 'main.go', 'refactor/')
    await page.waitForSelector('.cm-content', { timeout: 10000 })
    let summary = ''
    for (let attempt = 0; attempt < 24 && !SUMMARY[lang].test(summary); attempt++) {
      await L.sleep(2500)
      must(await cursorOn(page, 7, 'Greet'), 'Greet not on line 7')
      await page.keyboard.down('Shift')
      await page.keyboard.press('F12')
      await page.keyboard.up('Shift')
      summary = await page
        .waitForFunction(() => document.querySelector('.list .summary')?.innerText ?? '', { timeout: 6000 })
        .then((handle) => handle.jsonValue())
        .catch(() => '')
    }
    must(SUMMARY[lang].test(summary), `summary = "${summary}"`)
    const tab = await page.evaluate((name) => [...document.querySelectorAll('[role=tab]')].find((el) => el.textContent.includes(name))?.getAttribute('aria-selected'), REFERENCES_TAB[lang])
    must(tab === 'true', 'the References tab is not selected')
    const files = await page.$$eval('.list .file .name', (els) => els.map((el) => el.textContent.trim()))
    must(files.join(',') === 'helper.go,main.go', `files = ${files}`)
    await L.shot(page, `refactor-${lang}-03-references`)
    await page.evaluate(() => document.querySelector('.list .group .ref')?.click()) // the first one: helper.go
    await page.waitForFunction(() => document.querySelector('.crumbs b')?.textContent === 'helper.go', { timeout: 10000 })
    const line = await U.cursorLine(page)
    must(line === 4, `the cursor is on line ${line}, want 4 (the declaration)`)
    return `${summary}; click opened helper.go at line ${line}`
  })

  return list
}
