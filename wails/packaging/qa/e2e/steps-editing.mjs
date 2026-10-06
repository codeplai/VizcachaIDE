// Editing features: replace in the Search panel (an open file changes in its buffer, a closed one
// on disk), Quick open (Ctrl+P) and the snippets of the completion list. The replace steps write
// their own folder (qa/editing: main.go and helper.go, both mention `palabraUnica`, a word no other
// phase uses), so other phases' files never change the counts. Registered in qa.mjs as `editing`.
import fs from 'node:fs'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { must, pageOf } from './interactions.mjs'

const DESCRIPTIONS = {
  en: 'for: repeat a block a number of times',
  es: 'for: repite un bloque un número de veces'
}
const SUMMARY = { en: /2 results in 2 files/, es: /2 resultados en 2 archivos/ }

const ctrl = async (page, ...keys) => {
  await page.keyboard.down('Control')
  for (const key of keys) await page.keyboard.down(key)
  for (const key of [...keys].reverse()) await page.keyboard.up(key)
  await page.keyboard.up('Control')
}

export const editingSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const folder = U.projectFile('qa', 'editing')
  const helper = U.projectFile('qa', 'editing', 'helper.go')
  const original = () => fs.readFileSync(helper, 'utf8')
  const writeFolder = () => {
    fs.mkdirSync(folder, { recursive: true })
    fs.writeFileSync(U.projectFile('qa', 'editing', 'main.go'), 'package main\n\nfunc main() { println(palabraUnica()) }\n')
    fs.writeFileSync(helper, 'package main\n\nfunc palabraUnica() string { return "ok" }\n')
  }
  const status = () => page.evaluate(() => document.querySelector('.side .status')?.innerText ?? '')
  const panelInputs = () => page.$$('.side input.input')

  const search = async (query) => {
    await page.evaluate(() => document.querySelector('.cm-content')?.focus() ?? document.body.focus())
    await ctrl(page, 'Shift', 'KeyF')
    await page.waitForSelector('.side input.input', { timeout: 10000 })
    const [field] = await panelInputs()
    await field.click({ clickCount: 3 })
    await page.keyboard.type(query)
    await page.keyboard.press('Enter')
    await page.waitForFunction(() => /\d/.test(document.querySelector('.side .status')?.innerText ?? ''), { timeout: 15000 })
  }

  const openReplace = async (text) => {
    if ((await panelInputs()).length < 2) await page.click('.side button.chev')
    const [, field] = await panelInputs()
    await field.click({ clickCount: 3 })
    await page.keyboard.type(text)
  }

  add('replace-disk', 'Search > Replace all: confirmation names matches and files; a closed file changes on disk', async () => {
    writeFolder()
    await U.reload(page, ctx.url)
    await search('palabraUnica')
    must(SUMMARY[lang].test(await status()), `status = ${await status()}`)
    await openReplace('saludo')
    await page.click('.side .all')
    await page.waitForSelector('[role=alertdialog]', { timeout: 10000 })
    const message = await page.$eval('[role=alertdialog]', (el) => el.innerText)
    must(/2/.test(message), `confirmation = ${message}`)
    await L.shot(page, `${lang}-editing-01-confirm`)
    await page.evaluate(() => document.querySelector('[role=alertdialog] .dlg-button.primary').click())
    await page.waitForFunction(() => !document.querySelector('[role=alertdialog]'), { timeout: 10000 })
    await L.sleep(800)
    const onDisk = original()
    must(onDisk.includes('saludo') && !onDisk.includes('palabraUnica'), `helper.go on disk: ${onDisk}`)
    return 'helper.go on disk now says saludo'
  })

  add('replace-open', 'Replace in an open file edits the buffer: unsaved mark, disk untouched, Ctrl+Z undoes it', async () => {
    writeFolder()
    await U.reload(page, ctx.url)
    await U.openByName(page, 'helper.go', 'qa/editing')
    await search('palabraUnica')
    await openReplace('hola')
    // The per-match button of helper.go's group (main.go's group may come first).
    await page.evaluate(() => {
      const group = [...document.querySelectorAll('.side .file')].find(
        (g) => g.querySelector('.file-name')?.textContent.trim() === 'helper.go'
      )
      group.querySelector('li.match .mini').click()
    })
    await L.sleep(800)
    const text = await U.editorText(page)
    must(text.includes('hola'), `editor = ${text}`)
    must(await page.$('.tab .dot, [role=tab] .dot'), 'no unsaved mark on the tab')
    must(original().includes('palabraUnica'), 'the file on disk changed')
    await page.evaluate(() => document.querySelector('.cm-content').focus())
    await ctrl(page, 'KeyZ')
    await L.sleep(400)
    must((await U.editorText(page)).includes('palabraUnica'), 'Ctrl+Z did not undo the replacement')
    await L.shot(page, `${lang}-editing-02-open`)
    return 'buffer changed, unsaved, undone'
  })

  add('quick-open', 'Ctrl+P: fuzzy list with matched letters, Enter opens, Escape closes', async () => {
    await U.reload(page, ctx.url)
    await page.evaluate(() => document.querySelector('.cm-content')?.focus() ?? document.body.focus())
    await ctrl(page, 'KeyP')
    await page.waitForSelector('[role=combobox]', { timeout: 10000 })
    await page.keyboard.type('hlpr')
    await page.waitForSelector('[role=option] mark', { timeout: 10000 })
    const first = await page.$eval('[role=option]', (el) => el.innerText.replace(/\s+/g, ' '))
    must(first.includes('helper.go'), `first = ${first}`)
    await L.shot(page, `${lang}-editing-03-quick-open`)
    await page.keyboard.press('Enter')
    await page.waitForFunction(() => document.querySelector('.crumbs b')?.textContent === 'helper.go', { timeout: 10000 })
    await ctrl(page, 'KeyP')
    await page.waitForSelector('[role=combobox]', { timeout: 10000 })
    await page.keyboard.press('Escape')
    await page.waitForFunction(() => !document.querySelector('[role=combobox]'), { timeout: 5000 })
    return first
  })

  add('snippet', 'Typing `for` offers a snippet with its description; Enter inserts it with tab stops', async () => {
    await U.reload(page, ctx.url)
    await U.openByName(page, 'main.go', 'qa/mod')
    await U.setEditorText(page, 'package main\n\nfunc main() {\n\t')
    await page.evaluate(() => {
      const view = document.querySelector('.cm-content').cmTile.view
      view.dispatch({ selection: { anchor: view.state.doc.length } })
      view.focus()
    })
    await page.keyboard.type('for')
    await page.waitForSelector('.cm-tooltip-autocomplete .cm-completionIcon-snippet', { timeout: 15000 })
    const info = await page.waitForFunction(() => document.querySelector('.cm-completionInfo')?.innerText || false, { timeout: 10000 })
    must((await info.jsonValue()).includes(DESCRIPTIONS[lang]), `description = ${await info.jsonValue()}`)
    await L.shot(page, `${lang}-editing-04-snippet`)
    await page.keyboard.press('Enter')
    await L.sleep(500)
    const text = await U.editorText(page)
    must(text.includes('for i := 0; i < n; i++ {'), `editor = ${text}`)
    await page.keyboard.type('k')
    must((await U.editorText(page)).includes('for k := 0; k < n; k++ {'), 'the tab stop did not mirror the name')
    return 'snippet inserted, first stop selected'
  })

  add('restore', 'Leaves the QA project as it was', async () => {
    // The IDE still watches the folder (removing it fails with EPERM): put the files back instead.
    writeFolder()
    return 'qa/editing files restored'
  })

  return list
}
