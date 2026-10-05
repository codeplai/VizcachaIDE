// Python checks of M1 (docs/PLAN_PYTHON.md section 10): run with input(), an explained error, live
// problems from pyflakes, the debugger (variables and keyboard input while debugging) and the
// console. The session must set Settings.toolPaths.python to a Python with debugpy, pylsp, pyflakes
// and ruff (qa.mjs uses the development venv wails/.venv-py312).
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'
import { answer, must, pageOf, press, toggleBreakpoint } from './interactions.mjs'

const NAME_ERROR = { en: 'Python does not know the name "nombre"', es: 'Python no conoce el nombre "nombre"' }

const PROGRAMS = {
  'hola.py': 'nombre = input("Nombre: ")\nprint("Hola, " + nombre)\n',
  'error.py': 'print("antes")\nprint(nombre)\n',
  'factorial.py': 'def factorial(n):\n    resultado = 1\n    for i in range(2, n + 1):\n        resultado = resultado * i\n    return resultado\n\nprint(factorial(5))\n',
  'entrada.py': 'color = input("Color: ")\nprint("Te gusta el " + color)\n'
}

export const pythonSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const folder = path.join(L.PROJECT, 'qa', 'py')
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const chip = () => page.evaluate(() => document.querySelector('.guide-h .chip')?.textContent.trim() ?? '')
  const open = async (name) => {
    await L.go(page, 'FilesService', 'ListTree', folder)
    await U.reload(page, ctx.url)
    await U.openByName(page, name)
    await page.waitForSelector('.cm-content', { timeout: 10000 })
  }

  add('py-run', 'hola.py: F5 runs it in a terminal; input() reads what is typed in Output', async () => {
    fs.mkdirSync(folder, { recursive: true })
    for (const [name, text] of Object.entries(PROGRAMS)) fs.writeFileSync(path.join(folder, name), text)
    await open('hola.py')
    await U.run(page)
    await U.waitOutput(page, 'Nombre:')
    await answer(page, 'Ana')
    await U.waitOutput(page, 'Hola, Ana')
    await U.waitOutput(page, T.finished)
    const text = await U.outputText(page)
    must(text.split('Ana').length - 1 >= 2 || text.includes('Nombre: Ana'), `typed text not shown once: ${text}`)
    await L.shot(page, `${lang}-py-01-run`)
    return 'Nombre: Ana -> Hola, Ana'
  })

  add('py-error', 'error.py: the NameError is explained by the Assistant in the IDE language', async () => {
    await open('error.py')
    await U.run(page)
    await page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout: 30000 }, NAME_ERROR[lang])
    await L.shot(page, `${lang}-py-02-error`)
    return `card "${NAME_ERROR[lang]}"`
  })

  add('py-live', 'error.py: pyflakes underlines the undefined name while editing', async () => {
    await open('error.py')
    await page.waitForFunction(
      () => [...document.querySelectorAll('.cm-lintRange-error, .cm-lintRange-warning')].some((e) => e.textContent.includes('nombre')),
      { timeout: 45000 }
    )
    await L.shot(page, `${lang}-py-03-live`)
    return 'underline on "nombre"'
  })

  add('py-debug', 'factorial.py: breakpoint on line 4, Debug, variables n/resultado/i, Stop debugging', async () => {
    await open('factorial.py')
    await toggleBreakpoint(page, 4)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 90000 }, `${T.paused} 4`)
    await page.waitForSelector('.guide .var', { timeout: 20000 })
    const vars = await page.$$eval('.guide .var', (els) => els.map((e) => [e.querySelector('.n')?.textContent.trim(), e.querySelector('.v')?.textContent.trim()]))
    const map = Object.fromEntries(vars)
    must(map.n === '5' && map.resultado === '1' && map.i === '2', `variables = ${JSON.stringify(map)}`)
    must(!vars.some(([name]) => /^__|special variables/.test(name ?? '')), `hidden variables shown: ${JSON.stringify(map)}`)
    await L.shot(page, `${lang}-py-04-debug`)
    const where = await chip()
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    await page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout: 30000 }, T.run)
    return `${where}; ${JSON.stringify(map)}; stopped`
  })

  add('py-debug-input', 'entrada.py: while debugging, input() reads what is typed in Output', async () => {
    await open('entrada.py')
    await press(page, '.titlebar .btn', T.debug)
    await U.waitOutput(page, 'Color:', 90000)
    await answer(page, 'verde')
    await U.waitOutput(page, 'Te gusta el verde', 30000)
    await L.shot(page, `${lang}-py-05-debug-input`)
    await page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout: 30000 }, T.run)
    return 'Color: verde -> Te gusta el verde'
  })

  add('py-console', 'Console: >>> prompt, 2 + 2 = 4 and x = 5 is remembered', async () => {
    await open('hola.py')
    await press(page, '[role=tab]', lang === 'es' ? 'Consola' : 'Console')
    await page.waitForSelector('.console textarea', { timeout: 10000 })
    const prompt = await page.$eval('.console .bar .prompt', (e) => e.textContent.trim())
    must(prompt === '>>>', `prompt = ${prompt}`)
    for (const code of ['2 + 2', 'x = 5', 'x * 2']) {
      await page.type('.console textarea', code)
      await page.keyboard.press('Enter')
      await L.sleep(1500)
    }
    const results = await page.$$eval('.console .result', (els) => els.map((e) => e.textContent.trim()))
    must(results.includes('4') && results.includes('10'), `results = ${results}`)
    await L.shot(page, `${lang}-py-06-console`)
    return `prompt >>>; results ${results.join(', ')}`
  })

  add('py-packages-search', 'Packages: typing "nump" lists PyPI packages; choosing numpy installs it', async () => {
    await open('hola.py')
    await page.click('.more-trigger')
    await press(page, '.menu-item', lang === 'es' ? 'Paquetes' : 'Packages')
    await page.waitForSelector('#module-package', { timeout: 10000 })
    await page.type('#module-package', 'nump')
    await page.waitForSelector('[role=listbox] [role=option]', { timeout: 30000 })
    const options = await page.$$eval('[role=listbox] [role=option] .name', (els) => els.map((e) => e.textContent.trim()))
    must(options[0] === 'numpy', `options = ${options.join(', ')}`)
    await L.shot(page, `${lang}-py-07-packages-search`)
    await page.click('[role=listbox] [role=option]')
    must((await page.$eval('#module-package', (e) => e.value)) === 'numpy', 'the chosen name is not in the field')
    await page.keyboard.press('Enter')
    await page.waitForFunction(() => /✓/.test(document.querySelector('[role=dialog]')?.innerText ?? ''), { timeout: 300000 })
    const dialog = await page.$eval('[role=dialog]', (e) => e.innerText)
    must(!/\[notice\]/.test(dialog), 'pip still prints its version notice')
    await page.keyboard.press('Escape')
    // The pip command must not be checked like a program (ruff reported "os error 2" on it).
    await L.sleep(6000)
    await press(page, '[role=tab]', lang === 'es' ? 'Problemas' : 'Problems')
    await L.sleep(500)
    const problems = await page.$$eval('button.problem', (els) => els.map((e) => e.innerText))
    must(!problems.some((p) => /pip|os error/.test(p)), `problems after installing: ${problems.join(' | ')}`)
    return `options ${options.slice(0, 4).join(', ')}…; numpy installed; no problems`
  })

  return list
}
