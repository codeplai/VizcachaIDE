// File > New project in every language: the dialog (Ctrl+Shift+N lists the languages and keeps
// Create disabled until a location is chosen) and, for Go, Python, C++ and Rust, a project created
// by the backend the dialog calls (the native folder picker cannot be driven headless), opened in
// the Files panel and run with F5: the template asks for a name and greets.
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'
import { answer, must, pageOf } from './interactions.mjs'

const MAIN_FILES = { go: 'main.go', python: 'main.py', cpp: 'main.cpp', rust: 'main.rs' }

export const newProjectSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const location = path.join(L.PROJECT, 'qa', 'nuevos')
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })

  add('new-dialog', 'Ctrl+Shift+N opens New project: the four languages, Create disabled without a location', async () => {
    fs.rmSync(location, { recursive: true, force: true })
    fs.mkdirSync(location, { recursive: true })
    await U.reload(page, ctx.url)
    await page.evaluate(() => document.querySelector('.cm-content')?.focus() ?? document.body.focus())
    await page.keyboard.down('Control')
    await page.keyboard.down('Shift')
    await page.keyboard.press('KeyN')
    await page.keyboard.up('Shift')
    await page.keyboard.up('Control')
    await page.waitForSelector('[role=dialog] select', { timeout: 10000 })
    const languages = await page.$$eval('[role=dialog] select option', (els) => els.map((e) => e.value))
    await page.type('[role=dialog] input[type=text], [role=dialog] input:not([readonly])', 'Mi Tienda')
    const disabled = await page.evaluate(() => {
      const buttons = [...document.querySelectorAll('[role=dialog] button')]
      return buttons.at(-1)?.disabled ?? null
    })
    await L.shot(page, `${lang}-new-01-dialog`)
    await page.keyboard.press('Escape')
    for (const id of ['go', 'python', 'cpp', 'rust']) must(languages.includes(id), `languages = ${languages}`)
    must(disabled === true, `Create enabled without a location (${disabled})`)
    return `languages ${languages.join(', ')}; Create disabled`
  })

  for (const codeLanguage of ['go', 'python', 'cpp', 'rust']) {
    add(`new-${codeLanguage}`, `${codeLanguage}: a new project "Mi Proyecto Ñandú" opens and runs with F5`, async () => {
      const name = `Mi Proyecto Ñandú ${codeLanguage}`
      const created = await L.go(page, 'ProjectsService', 'Create', codeLanguage, location, name)
      must(created.root === path.join(location, name), `root = ${created.root}`)
      must(fs.existsSync(created.mainFile), `no ${created.mainFile}`)
      await L.go(page, 'FilesService', 'ListTree', created.root)
      await U.reload(page, ctx.url)
      await U.openByName(page, MAIN_FILES[codeLanguage], `${name}/`)
      await U.run(page)
      await U.waitOutput(page, '¿Cómo te llamas?', 300000)
      await answer(page, 'Ñandú') // accents typed and printed (Windows console code page)
      await U.waitOutput(page, 'Hola, Ñandú', 60000)
      await U.waitOutput(page, T.finished, 60000)
      await L.shot(page, `${lang}-new-${codeLanguage}`)
      const files = fs.readdirSync(created.root, { recursive: true }).filter((f) => !String(f).startsWith('target'))
      return `${files.join(', ')} -> Hola, Ñandú`
    })
  }
  add('close-folder', 'The Files panel closes the open folder: its tabs close and the panel is empty', async () => {
    const name = 'Para Cerrar'
    const created = await L.go(page, 'ProjectsService', 'Create', 'python', location, name)
    await L.go(page, 'FilesService', 'ListTree', created.root)
    await U.reload(page, ctx.url)
    await U.openByName(page, 'main.py', `${name}/`)
    const label = lang === 'es' ? 'Cerrar carpeta' : 'Close folder'
    await page.click(`.side.files button.tool[aria-label="${label}"]`)
    await page.waitForFunction(() => !document.querySelector('.tree'), { timeout: 10000 })
    const tabs = await page.$$eval('.tabs [role=tab]', (els) => els.map((e) => e.textContent.trim()))
    must(!tabs.some((tab) => tab.includes('main.py')), `tabs still open: ${tabs}`)
    await L.shot(page, `${lang}-new-close-folder`)
    await U.reload(page, ctx.url)
    must(!(await page.$('.tree')), 'the closed folder came back after a reload')
    return 'tree empty, main.py closed, not reopened after a reload'
  })

  return list
}
