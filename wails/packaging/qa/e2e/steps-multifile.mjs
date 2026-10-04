// Multi-file projects with an external library, in every language (the user's experiment after
// M3): main calls a function in a second file that calls a third one, and a fourth file uses an
// external library added with the package manager (Go modules, pip, cargo) or copied next to the
// code (a header-only C++ library). For each language: the library is added, the program runs,
// Ctrl+click goes to the definition in the other file, hover documents the library's function, a
// breakpoint in the third file shows the whole call stack, and an error in the third file is
// listed in Problems and opens that file.
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'
import { must, pageOf, press, toggleBreakpoint } from './interactions.mjs'
import { MULTIFILE_PROJECTS } from './multifile-projects.mjs'

const fileOf = (relative) => relative.split('/').at(-1)

/** Where the mouse goes for a word of a line (its third character). */
const wordAt = (page, line, word) =>
  page.evaluate(
    (n, w) => {
      const view = document.querySelector('.cm-content').cmTile.view
      const text = view.state.doc.line(n)
      const at = text.text.indexOf(w)
      if (at < 0) return null
      const c = view.coordsAtPos(text.from + at + 2)
      return { x: c.left, y: (c.top + c.bottom) / 2 }
    },
    line,
    word
  )

const activeFile = (page) => page.evaluate(() => document.querySelector('.crumbs b')?.textContent ?? '')

export const multifileSteps = (ctx, lang, codeLanguage) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const project = MULTIFILE_PROJECTS[codeLanguage]
  const root = path.join(L.PROJECT, 'qa', 'multi', project.dir)
  const list = []
  const add = (id, title, run) => list.push({ id: `${codeLanguage}-${id}`, title, run })
  const folderOf = (relative) => `multi/${project.dir}/${relative}`.split('/').slice(0, -1).join('/') + '/'
  /** Opens a file of the project in a fresh window (the tree shows the project). */
  const open = async (relative) => {
    await L.go(page, 'FilesService', 'ListTree', root)
    await U.reload(page, ctx.url)
    await openTab(relative)
  }
  /** Opens a file without reloading (breakpoints set in other files are kept). */
  const openTab = async (relative) => {
    await U.openByName(page, fileOf(relative), folderOf(relative))
    await page.waitForSelector('.cm-content', { timeout: 10000 })
  }
  const runEnabled = () =>
    page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout: 300000 }, T.run)

  add('library', `${codeLanguage}: write the 4 files and add the external library`, async () => {
    fs.rmSync(root, { recursive: true, force: true })
    for (const [name, text] of Object.entries(project.files)) {
      fs.mkdirSync(path.join(root, path.dirname(name)), { recursive: true })
      fs.writeFileSync(path.join(root, name), text)
    }
    const library = project.library
    if (library.download) {
      const cached = path.join(L.WQ, library.file)
      if (!fs.existsSync(cached)) fs.writeFileSync(cached, Buffer.from(await (await fetch(library.download)).arrayBuffer()))
      fs.copyFileSync(cached, path.join(root, library.file))
      return `${library.file} copied (${fs.statSync(cached).size} bytes)`
    }
    await open(project.main)
    await L.go(page, 'PackagesService', 'Add', library.codeLanguage, root, library.name)
    await L.sleep(2000) // the command starts as a run: wait for it to end
    await runEnabled()
    const output = await U.outputText(page)
    if (library.added) {
      const manifest = fs.readFileSync(path.join(root, library.added[0]), 'utf8')
      must(manifest.includes(library.added[1]), `${library.added[0]} has no ${library.added[1]}:\n${output}`)
    }
    return `${library.name} added`
  })

  add('run', `${codeLanguage}: F5 runs main, which uses the other files and the library`, async () => {
    await open(project.main)
    await U.run(page)
    for (const line of project.output) await U.waitOutput(page, line, 300000)
    await runEnabled()
    await L.shot(page, `multi-${codeLanguage}-01-run`)
    return project.output.join(' / ')
  })

  add('definition', `${codeLanguage}: Ctrl+click on a function of another file opens that file`, async () => {
    await open(project.main)
    const where = project.definition
    let reached = ''
    for (let attempt = 0; attempt < 12 && !where.files.includes(reached); attempt++) {
      await L.sleep(5000) // the code helper indexes the project (rust-analyzer takes seconds)
      const at = await wordAt(page, where.line, where.word)
      must(at, `"${where.word}" not on line ${where.line}`)
      await page.keyboard.down('Control')
      await page.mouse.click(at.x, at.y)
      await page.keyboard.up('Control')
      await L.sleep(1500)
      reached = await activeFile(page)
    }
    must(where.files.includes(reached), `Ctrl+click on ${where.word} opened ${reached}`)
    return `${where.word} -> ${reached}`
  })

  add('external', `${codeLanguage}: hover documents the external library's function`, async () => {
    const where = project.external
    await open(where.file)
    let hover = ''
    for (let attempt = 0; attempt < 12 && !hover.includes(where.word); attempt++) {
      await L.sleep(5000)
      const at = await wordAt(page, where.line, where.word)
      must(at, `"${where.word}" not on line ${where.line}`)
      await page.mouse.move(5, 5)
      await page.mouse.move(at.x, at.y)
      hover = await page
        .waitForSelector('.cm-tooltip-hover', { timeout: 8000 })
        .then((e) => e.evaluate((n) => n.innerText))
        .catch(() => '')
    }
    must(hover.includes(where.word), `hover on ${where.word} = ${hover}`)
    await L.shot(page, `multi-${codeLanguage}-02-external`)
    return hover.split('\n')[0].slice(0, 120)
  })

  add('debug', `${codeLanguage}: a breakpoint in the third file shows main -> second -> third`, async () => {
    await open(project.breakpoint.file)
    await toggleBreakpoint(page, project.breakpoint.line)
    await openTab(project.main)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 300000 }, `${T.paused} ${project.breakpoint.line}`)
    await page.waitForSelector('ul.stack button', { timeout: 30000 })
    const frames = await page.$$eval('ul.stack button', (els) => els.map((e) => e.textContent.trim()))
    const vars = await page.$$eval('.guide .var', (els) => els.map((e) => [e.querySelector('.n')?.textContent.trim(), e.querySelector('.v')?.textContent.trim()]))
    const map = Object.fromEntries(vars)
    await L.shot(page, `multi-${codeLanguage}-03-debug`)
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    await runEnabled()
    for (const file of project.stack) must(frames.some((f) => f.includes(file)), `no ${file} in the stack: ${frames.join(' | ')}`)
    must(map.monto === '400' && map.impuesto === '40', `variables = ${JSON.stringify(map)}`)
    return `${frames.join(' | ')}; monto=400 impuesto=40`
  })

  add('error', `${codeLanguage}: an error in the third file is listed in Problems and opens that file`, async () => {
    const file = path.join(root, project.secondary.file)
    const original = fs.readFileSync(file, 'utf8')
    fs.writeFileSync(file, original.replace(project.secondary.from, project.secondary.to))
    try {
      await open(project.main)
      await U.run(page)
      await runEnabled()
      await press(page, '[role=tab]', lang === 'es' ? 'Problemas' : 'Problems')
      const name = fileOf(project.secondary.file)
      await page.waitForFunction((n) => [...document.querySelectorAll('button.problem')].some((b) => b.innerText.includes(n)), { timeout: 60000 }, name)
      await L.shot(page, `multi-${codeLanguage}-04-error`)
      const problem = await page.evaluateHandle((n) => [...document.querySelectorAll('button.problem')].find((b) => b.innerText.includes(n)), name)
      const text = await problem.evaluate((b) => b.innerText.replace(/\s+/g, ' '))
      await problem.asElement().click()
      await page.waitForFunction((n) => document.querySelector('.crumbs b')?.textContent === n, { timeout: 10000 }, name)
      return `${text.slice(0, 140)} -> opened ${name}`
    } finally {
      fs.writeFileSync(file, original)
    }
  })

  return list
}
