// C++ checks of M2 (docs/PLAN_CPP.md section 10): compile and run with std::cin, an explained
// compile error, an explained crash, live problems from clangd, Build from the More menu, the
// debugger (variables and keyboard input while debugging) and format on save. The session must set
// Settings.toolPaths.cxx to a clang++ with lldb-dap, clangd and clang-format next to it (qa.mjs uses
// the development llvm-mingw of docs/PLAN_CPP.md section 0).
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'
import { answer, must, pageOf, press, toggleBreakpoint } from './interactions.mjs'

const UNDECLARED = { en: 'The compiler does not know the name "totl"', es: 'El compilador no conoce el nombre "totl"' }
const SEGFAULT = { en: 'The program touched memory it is not allowed to use', es: 'El programa tocó memoria que no puede usar' }
const BUILD = { en: 'Build', es: 'Compilar' }

const PROGRAMS = {
  'hola/main.cpp': '#include <iostream>\n#include <string>\nusing namespace std;\n\nint main() {\n    string nombre;\n    cout << "Nombre: ";\n    cin >> nombre;\n    cout << "Hola, " << nombre << endl;\n    return 0;\n}\n',
  'error/main.cpp': '#include <iostream>\n\nint main() {\n    int total = 1;\n    std::cout << totl << std::endl;\n    return 0;\n}\n',
  'caida/main.cpp': '#include <iostream>\n\nint main() {\n    std::cout << "antes de caer" << std::endl;\n    int* p = nullptr;\n    *p = 42;\n    return 0;\n}\n',
  'factorial/main.cpp': '#include <iostream>\n\nint factorial(int n) {\n    int resultado = 1;\n    for (int i = 2; i <= n; i++) {\n        resultado = resultado * i;\n    }\n    return resultado;\n}\n\nint main() {\n    std::cout << factorial(5) << std::endl;\n    return 0;\n}\n',
  'entrada/main.cpp': '#include <iostream>\n#include <string>\n\nint main() {\n    std::string color;\n    std::cout << "Color: ";\n    std::cin >> color;\n    std::cout << "Te gusta el " << color << std::endl;\n    return 0;\n}\n',
  'formato/main.cpp': '#include <iostream>\nint main(){\nstd::cout<<"hola"<<std::endl;\n  return 0;}\n'
}

export const cppSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const folder = path.join(L.PROJECT, 'qa', 'cpp')
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const open = async (program) => {
    const dir = path.join(folder, path.dirname(program))
    await L.go(page, 'FilesService', 'ListTree', dir)
    await U.reload(page, ctx.url)
    await U.openByName(page, path.basename(program))
    await page.waitForSelector('.cm-content', { timeout: 10000 })
  }
  const runEnabled = () =>
    page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout: 30000 }, T.run)

  add('cpp-run', 'hola/main.cpp: F5 compiles and runs it in a terminal; std::cin reads what is typed in Output', async () => {
    fs.rmSync(folder, { recursive: true, force: true })
    for (const [name, text] of Object.entries(PROGRAMS)) {
      fs.mkdirSync(path.join(folder, path.dirname(name)), { recursive: true })
      fs.writeFileSync(path.join(folder, name), text)
    }
    await open('hola/main.cpp')
    await U.run(page)
    await U.waitOutput(page, 'Nombre:', 90000)
    await answer(page, 'Ana')
    await U.waitOutput(page, 'Hola, Ana')
    await U.waitOutput(page, T.finished)
    await L.shot(page, `${lang}-cpp-01-run`)
    return 'Nombre: Ana -> Hola, Ana'
  })

  add('cpp-error', 'error/main.cpp: the undeclared name is explained by the Assistant in the IDE language', async () => {
    await open('error/main.cpp')
    await U.run(page)
    await page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout: 60000 }, UNDECLARED[lang])
    await L.shot(page, `${lang}-cpp-02-error`)
    return `card "${UNDECLARED[lang]}"`
  })

  add('cpp-crash', 'caida/main.cpp: the null pointer ends with "Segmentation fault", explained', async () => {
    await open('caida/main.cpp')
    await U.run(page)
    await U.waitOutput(page, 'Segmentation fault', 90000)
    const text = await U.outputText(page)
    must(text.indexOf('antes de caer') < text.indexOf('Segmentation fault'), `crash line order: ${text}`)
    await page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout: 30000 }, SEGFAULT[lang])
    await L.shot(page, `${lang}-cpp-03-crash`)
    return `"Segmentation fault" + card "${SEGFAULT[lang]}"`
  })

  add('cpp-live', 'error/main.cpp: clangd underlines the undeclared name while editing', async () => {
    await open('error/main.cpp')
    await page.waitForFunction(
      () => [...document.querySelectorAll('.cm-lintRange-error')].some((e) => e.textContent.includes('totl')),
      { timeout: 60000 }
    )
    await L.shot(page, `${lang}-cpp-04-live`)
    return 'underline on "totl"'
  })

  add('cpp-build', 'hola/main.cpp: More > Build leaves main.exe next to the source without running it', async () => {
    const exe = path.join(folder, 'hola', process.platform === 'win32' ? 'main.exe' : 'main')
    fs.rmSync(exe, { force: true })
    await open('hola/main.cpp')
    await page.click('.more-trigger')
    await press(page, '.menu-item', BUILD[lang])
    for (let i = 0; i < 90 && !fs.existsSync(exe); i++) await L.sleep(1000)
    must(fs.existsSync(exe), `${exe} was not built`)
    await runEnabled()
    must(!(await U.outputText(page)).includes('Nombre:'), 'Build ran the program')
    return path.basename(exe)
  })

  add('cpp-debug', 'factorial/main.cpp: breakpoint on line 6, Debug, variables n/resultado/i, Stop debugging', async () => {
    await open('factorial/main.cpp')
    await toggleBreakpoint(page, 6)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 120000 }, `${T.paused} 6`)
    await page.waitForSelector('.guide .var', { timeout: 20000 })
    const vars = await page.$$eval('.guide .var', (els) => els.map((e) => [e.querySelector('.n')?.textContent.trim(), e.querySelector('.v')?.textContent.trim()]))
    const map = Object.fromEntries(vars)
    must(map.n === '5' && map.resultado === '1' && map.i === '2', `variables = ${JSON.stringify(map)}`)
    await L.shot(page, `${lang}-cpp-05-debug`)
    const frames = await page.evaluate(() => document.querySelector('.guide')?.innerText ?? '')
    must(!/__tmainCRTStartup|BaseThreadInitThunk/.test(frames), 'C runtime frames shown')
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    await runEnabled()
    return `${JSON.stringify(map)}; stopped`
  })

  add('cpp-debug-input', 'entrada/main.cpp: while debugging, std::cin reads what is typed in Output', async () => {
    await open('entrada/main.cpp')
    await press(page, '.titlebar .btn', T.debug)
    await U.waitOutput(page, 'Color:', 120000)
    await answer(page, 'verde')
    await U.waitOutput(page, 'Te gusta el verde', 30000)
    await L.shot(page, `${lang}-cpp-06-debug-input`)
    await runEnabled()
    return 'Color: verde -> Te gusta el verde'
  })

  add('cpp-format', 'formato/main.cpp: Ctrl+S formats it with clang-format (4 spaces)', async () => {
    const file = path.join(folder, 'formato', 'main.cpp')
    fs.writeFileSync(file, PROGRAMS['formato/main.cpp'])
    await open('formato/main.cpp')
    await page.evaluate(() => document.querySelector('.cm-content').focus())
    await page.keyboard.down('Control')
    await page.keyboard.press('s')
    await page.keyboard.up('Control')
    let saved = ''
    for (let i = 0; i < 30 && !saved.includes('    std::cout << "hola"'); i++) {
      await L.sleep(500)
      saved = fs.readFileSync(file, 'utf8')
    }
    must(saved.includes('    std::cout << "hola" << std::endl;'), `not formatted: ${JSON.stringify(saved)}`)
    return 'formatted with 4 spaces'
  })

  return list
}
