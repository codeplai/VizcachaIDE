// The Files panel: multi-select, copy / cut / paste, duplicate, drag and drop and delete of several
// entries, against the REAL backend (files on disk in a scratch folder of the project). Drag and
// drop is driven with synthetic DragEvents (headless Edge has no OS-level drag).
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import { must, pageOf, press } from './interactions.mjs'
import * as U from './ui.mjs'

const LAB = path.join(L.PROJECT, 'files-lab')
const TEXTS = {
  en: { duplicate: 'Duplicate', trash: 'Move to Recycle Bin', copySuffix: '(copy)' },
  es: { duplicate: 'Duplicar', trash: 'Mover a la Papelera', copySuffix: '(copia)' }
}

const seed = () => {
  fs.rmSync(LAB, { recursive: true, force: true })
  fs.mkdirSync(path.join(LAB, 'dest'), { recursive: true })
  for (const name of ['a', 'b', 'c', 'd', 'e', 'f']) fs.writeFileSync(path.join(LAB, `${name}.go`), `package lab // ${name}\n`)
}

export const filesSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const inLab = (...parts) => path.join(LAB, ...parts)
  const base = (name) => name.replace(/\.go$/, '')

  /** The row of the Files panel for a file or folder name. */
  const row = async (name) => {
    const handle = await page.evaluateHandle((n) => [...document.querySelectorAll('button.file')].find((b) => b.textContent.trim().endsWith(n) && (b.dataset.path ?? '').includes('files-lab')), name)
    const el = handle.asElement()
    must(el, `no row "${name}" in the Files panel`)
    await el.evaluate((b) => b.scrollIntoView({ block: 'center' }))
    return el
  }
  const click = async (name, modifier) => {
    const el = await row(name)
    if (modifier) await page.keyboard.down(modifier)
    await el.click()
    if (modifier) await page.keyboard.up(modifier)
  }
  const chord = async (key) => {
    await page.keyboard.down('Control')
    await page.keyboard.press(key)
    await page.keyboard.up('Control')
    await L.sleep(600)
  }
  const selected = () => page.$$eval('button.file.sel', (els) => els.map((e) => e.textContent.trim().replace(/^[▾▸]\s*/, '')))
  const waitFile = async (file, present = true) => {
    for (let i = 0; i < 40; i++) {
      if (fs.existsSync(file) === present) return
      await L.sleep(250)
    }
    throw new Error(`${file} ${present ? 'did not appear' : 'is still there'}`)
  }

  add('files-prepare', 'A scratch folder with six files and a "dest" folder shows in the Files panel', async () => {
    seed()
    await U.reload(page, ctx.url)
    await L.sleep(500)
    for (const name of ['a.go', 'f.go', 'dest']) await row(name)
    return 'rows a.go, f.go, dest found'
  })

  add('files-multiselect', 'Ctrl+click toggles, Shift+click selects a range, Escape clears', async () => {
    await click('a.go')
    await click('c.go', 'Control')
    const toggled = await selected()
    must(toggled.length === 2 && toggled.includes('a.go') && toggled.includes('c.go'), `ctrl: ${toggled}`)
    await click('a.go')
    await click('d.go', 'Shift')
    const range = await selected()
    must(['a.go', 'b.go', 'c.go', 'd.go'].every((n) => range.includes(n)) && !range.includes('e.go'), `shift: ${range}`)
    await page.keyboard.press('Escape')
    const none = await selected()
    must(none.length === 0, `escape left ${none}`)
    return `ctrl: ${toggled}; shift: ${range.length} rows; escape: ${none.length}`
  })

  add('files-copy-paste', 'Ctrl+C on a file and Ctrl+V on a folder copies it; a second paste never overwrites', async () => {
    await click('a.go')
    await chord('c')
    await click('dest')
    await chord('v')
    await waitFile(inLab('dest', 'a.go'))
    must(fs.readFileSync(inLab('dest', 'a.go'), 'utf8') === 'package lab // a\n', 'copied content differs')
    must(fs.existsSync(inLab('a.go')), 'the source disappeared')
    await chord('v')
    const copy = `${base('a.go')} ${T.copySuffix}.go`
    await waitFile(inLab('dest', copy))
    return `dest has ${fs.readdirSync(inLab('dest')).join(', ')}`
  })

  add('files-duplicate', 'Duplicate (context menu) copies in place with the language suffix', async () => {
    const el = await row('b.go')
    await el.click({ button: 'right' })
    await page.waitForSelector('[role=menu]', { timeout: 5000 })
    await press(page, '[role=menuitem]', T.duplicate)
    const copy = `b ${T.copySuffix}.go`
    await waitFile(inLab(copy))
    await L.shot(page, `${lang}-files-duplicate`)
    return copy
  })

  add('files-cut-paste-tab', 'Ctrl+X and Ctrl+V move an open file; its tab follows and saves to the new path', async () => {
    await (await row('c.go')).click({ clickCount: 2, delay: 30 })
    await page.waitForFunction(() => document.querySelector('.crumbs b')?.textContent === 'c.go', { timeout: 10000 })
    await U.setEditorText(page, 'package lab // moved\n')
    await click('c.go')
    await chord('x')
    await click('dest')
    await chord('v')
    await waitFile(inLab('dest', 'c.go'))
    await waitFile(inLab('c.go'), false)
    const title = await page.evaluate(() => document.querySelector('.crumbs b')?.textContent)
    must(title === 'c.go', `the tab lost the file: ${title}`)
    await page.click('.cm-content')
    await chord('s')
    await waitFile(inLab('dest', 'c.go'))
    for (let i = 0; i < 20 && !fs.readFileSync(inLab('dest', 'c.go'), 'utf8').includes('moved'); i++) await L.sleep(250)
    must(fs.readFileSync(inLab('dest', 'c.go'), 'utf8').includes('moved'), 'the save did not reach the new path')
    must(!fs.existsSync(inLab('c.go')), 'saving recreated the old path')
    return 'tab c.go follows to dest/ and Ctrl+S writes there'
  })

  add('files-drag-drop', 'Dragging a file onto a folder moves it and highlights the target; a folder cannot drop on itself', async () => {
    const outcome = await page.evaluate(async (names) => {
      const find = (n) => [...document.querySelectorAll('button.file')].find((b) => b.textContent.trim().endsWith(n) && (b.dataset.path ?? '').includes('files-lab'))
      const fire = (el, type, data) => el.dispatchEvent(new DragEvent(type, { bubbles: true, cancelable: true, dataTransfer: data }))
      const data = new DataTransfer()
      const [from, to] = [find(names[0]), find(names[1])]
      fire(from, 'dragstart', data)
      fire(to, 'dragover', data)
      await new Promise((resolve) => setTimeout(resolve, 100))
      const highlighted = to.classList.contains('drop')
      fire(to, 'drop', data)
      fire(from, 'dragend', data)
      return highlighted
    }, ['d.go', 'dest'])
    must(outcome, 'the drop target was not highlighted while dragging')
    await waitFile(inLab('dest', 'd.go'))
    await waitFile(inLab('d.go'), false)
    const refused = await page.evaluate(() => {
      const dest = [...document.querySelectorAll('button.file.dir')].find((b) => b.textContent.trim().endsWith('dest'))
      const data = new DataTransfer()
      dest.dispatchEvent(new DragEvent('dragstart', { bubbles: true, cancelable: true, dataTransfer: data }))
      const event = new DragEvent('dragover', { bubbles: true, cancelable: true, dataTransfer: data })
      dest.dispatchEvent(event)
      dest.dispatchEvent(new DragEvent('dragend', { bubbles: true, dataTransfer: data }))
      return event.defaultPrevented
    })
    must(!refused, 'a folder accepted a drop on itself')
    return 'd.go moved into dest/; dest refuses itself'
  })

  add('files-multi-delete', 'Delete on two selected files asks once with the count and trashes both', async () => {
    await click('e.go')
    await click('f.go', 'Control')
    await page.keyboard.press('Delete')
    await page.waitForSelector('[role=dialog], [role=alertdialog]', { timeout: 5000 })
    const text = await page.$eval('[role=dialog], [role=alertdialog]', (e) => e.innerText.replace(/\n+/g, ' | '))
    must(/\b2\b/.test(text), `the confirmation does not say 2: ${text}`)
    await L.shot(page, `${lang}-files-delete-many`)
    await L.clickText(page, '[role=dialog] button, [role=alertdialog] button', T.trash)
    await waitFile(inLab('e.go'), false)
    await waitFile(inLab('f.go'), false)
    return text.slice(0, 120)
  })

  return list
}
