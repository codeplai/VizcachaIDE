// C++ with CMake and vcpkg (M4, docs/PLAN_CPP_CMAKE.md): a new project is a CMake project and runs;
// the Packages dialog finds fmt in the offline index of vcpkg, installs it (minutes the first time,
// the binary cache makes it fast afterwards) and F5, Problems and the debugger work with it;
// removing it leaves vcpkg.json and CMakeLists.txt clean; and an old folder without CMakeLists.txt
// runs and gets one. The session needs Settings.toolPaths for the compiler, CMake, Ninja and vcpkg
// (L.DEV_CPP_TOOLS) and VIZCACHA_TEST_VCPKG_SEED (L.useDevVcpkgSeed).
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'
import { answer, must, pageOf, press, toggleBreakpoint } from './interactions.mjs'

const NAME = 'Ñandú'
const FILES = ['CMakeLists.txt', 'CMakePresets.json', 'vcpkg.json', '.gitignore', '.clang-format', 'main.cpp']
const FMT_PROGRAM = `#include <fmt/core.h>
#include <iostream>
#include <string>

int main() {
    std::string nombre;
    std::cout << "Nombre: " << std::flush;
    std::getline(std::cin, nombre);
    fmt::print("Hola desde fmt, {}!\\n", nombre);
    return 0;
}
`
const OLD = {
  'a.cpp': '#include <iostream>\n#include <string>\n\nstd::string saludo(const std::string& nombre);\n\nint main() {\n    std::string nombre;\n    std::cout << "Nombre: ";\n    std::getline(std::cin, nombre);\n    std::cout << saludo(nombre) << std::endl;\n    return 0;\n}\n',
  'b.cpp': '#include <string>\n\nstd::string saludo(const std::string& nombre) {\n    return "Hola desde b.cpp, " + nombre;\n}\n'
}
const INSTALL_TIMEOUT = 30 * 60 * 1000

export const cppCmakeSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const location = path.join(L.PROJECT, 'qa', 'cmake')
  const root = path.join(location, NAME)
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const open = async (folder, file) => {
    await L.go(page, 'FilesService', 'ListTree', folder)
    await U.reload(page, ctx.url)
    await U.openByName(page, file, `${path.basename(folder)}/`)
    await page.waitForSelector('.cm-content', { timeout: 10000 })
  }
  const runEnabled = (timeout = 300000) =>
    page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout }, T.run)
  const read = (file) => fs.readFileSync(path.join(root, file), 'utf8')
  const openPackages = async () => {
    await page.click('.more-trigger')
    await press(page, '.menu-item', lang === 'es' ? 'Paquetes' : 'Packages')
    await page.waitForSelector('#module-package', { timeout: 10000 })
  }
  // The dialog still shows the ✓ of an earlier run until the new command starts: wait for "Working…"
  // first, then for the ✓ without it.
  const dialogDone = async () => {
    await page.waitForFunction(() => /Working…|Trabajando…/.test(document.querySelector('[role=dialog]')?.innerText ?? ''), { timeout: 30000, polling: 100 })
    await page.waitForFunction(
      () => {
        const dialog = document.querySelector('[role=dialog]')?.innerText ?? ''
        return /✓/.test(dialog) && !/Working…|Trabajando…/.test(dialog)
      },
      { timeout: INSTALL_TIMEOUT, polling: 500 }
    )
  }
  const problems = async () => {
    await press(page, '[role=tab]', lang === 'es' ? 'Problemas' : 'Problems')
    await L.sleep(500)
    return page.$$eval('button.problem', (els) => els.map((e) => e.innerText.replace(/\s+/g, ' ')))
  }

  add('cmake-new', `New project C++ "${NAME}": the CMake files exist and F5 greets by name`, async () => {
    fs.rmSync(location, { recursive: true, force: true })
    fs.mkdirSync(location, { recursive: true })
    await U.reload(page, ctx.url)
    const created = await L.go(page, 'ProjectsService', 'Create', 'cpp', location, NAME)
    must(created.root === root, `root = ${created.root}`)
    for (const file of FILES) must(fs.existsSync(path.join(root, file)), `no ${file}`)
    must(!fs.existsSync(path.join(root, 'compile_flags.txt')), 'compile_flags.txt was created')
    await open(root, 'main.cpp')
    await U.run(page)
    await U.waitOutput(page, '¿Cómo te llamas?', 600000)
    await answer(page, NAME)
    await U.waitOutput(page, `Hola, ${NAME}`, 60000)
    await U.waitOutput(page, T.finished, 60000)
    must(fs.existsSync(path.join(root, 'build', 'compile_commands.json')), 'no build/compile_commands.json')
    await L.shot(page, `${lang}-cmake-01-new`)
    return `${FILES.join(', ')}; Hola, ${NAME}`
  })

  add('cmake-search-install', 'Packages: searching "fmt" lists vcpkg ports; choosing fmt installs it', async () => {
    await openPackages()
    await page.type('#module-package', 'fmt')
    await page.waitForSelector('[role=listbox] [role=option]', { timeout: 60000 })
    const options = await page.$$eval('[role=listbox] [role=option] .name', (els) => els.map((e) => e.textContent.trim()))
    must(options[0] === 'fmt', `options = ${options.join(', ')}`)
    await L.shot(page, `${lang}-cmake-02-search`)
    await page.click('[role=listbox] [role=option]')
    must((await page.$eval('#module-package', (e) => e.value)) === 'fmt', 'the chosen name is not in the field')
    await page.keyboard.press('Enter')
    await dialogDone()
    const notice = await page.$eval('[role=dialog]', (e) => e.innerText)
    await L.shot(page, `${lang}-cmake-03-installed`)
    await page.keyboard.press('Escape')
    const manifest = read('vcpkg.json')
    const block = read('CMakeLists.txt')
    must(manifest.includes('"fmt"'), `vcpkg.json = ${manifest}`)
    must(block.includes('find_package(fmt') && block.includes('fmt::fmt'), `CMakeLists.txt has no fmt block:\n${block}`)
    return `fmt installed; first-install notice shown: ${/first time|primera vez/i.test(notice)}`
  })

  add('cmake-use', 'main.cpp with fmt::print: F5 prints it and Problems is empty', async () => {
    fs.writeFileSync(path.join(root, 'main.cpp'), FMT_PROGRAM)
    await open(root, 'main.cpp')
    await U.run(page)
    await U.waitOutput(page, 'Nombre:', 300000)
    await answer(page, NAME)
    await U.waitOutput(page, `Hola desde fmt, ${NAME}!`, 300000)
    await U.waitOutput(page, T.finished, 60000)
    await L.sleep(6000) // clangd reads build/compile_commands.json and publishes its diagnostics
    const found = await problems()
    await L.shot(page, `${lang}-cmake-04-fmt`)
    must(found.length === 0, `problems: ${found.join(' | ')}`)
    return `Hola desde fmt, ${NAME}!; no problems`
  })

  add('cmake-debug', 'main.cpp with fmt: a breakpoint on the fmt::print line stops the debugger', async () => {
    await open(root, 'main.cpp')
    await toggleBreakpoint(page, 9)
    await press(page, '.titlebar .btn', T.debug)
    await U.waitOutput(page, 'Nombre:', 300000)
    await answer(page, NAME) // the program reads its name before it reaches the breakpoint
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 300000 }, `${T.paused} 9`)
    await page.waitForSelector('.guide .var', { timeout: 20000 })
    const vars = await page.$$eval('.guide .var', (els) => els.map((e) => [e.querySelector('.n')?.textContent.trim(), e.querySelector('.v')?.textContent.trim()]))
    await L.shot(page, `${lang}-cmake-05-debug`)
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    await runEnabled()
    must(vars.some(([name]) => /nombre$/.test(name ?? '')), `variables = ${JSON.stringify(vars)}`)
    return `stopped on line 9; ${vars.map(([n]) => n).join(', ')}`
  })

  add('cmake-remove', 'Packages: removing fmt leaves vcpkg.json and CMakeLists.txt clean', async () => {
    await open(root, 'main.cpp')
    await openPackages()
    await page.type('#module-package', 'fmt')
    await press(page, '[role=dialog] button', lang === 'es' ? 'Desinstalar' : 'Uninstall')
    await dialogDone()
    await page.keyboard.press('Escape')
    const manifest = JSON.parse(read('vcpkg.json'))
    const block = read('CMakeLists.txt')
    must(manifest.dependencies.length === 0, `vcpkg.json = ${JSON.stringify(manifest)}`)
    must(!block.includes('find_package(fmt') && !block.includes('fmt::fmt'), `CMakeLists.txt still has fmt:\n${block}`)
    return 'dependencies [] and no find_package(fmt)'
  })

  add('cmake-old-folder', 'An old folder with two .cpp and no CMakeLists.txt runs and gets one', async () => {
    const folder = path.join(location, 'carpeta vieja')
    fs.mkdirSync(folder, { recursive: true })
    for (const [name, text] of Object.entries(OLD)) fs.writeFileSync(path.join(folder, name), text)
    must(!fs.existsSync(path.join(folder, 'CMakeLists.txt')), 'the folder already had a CMakeLists.txt')
    await open(folder, 'a.cpp')
    await U.run(page)
    await U.waitOutput(page, 'Nombre:', 300000)
    await answer(page, 'Ana')
    await U.waitOutput(page, 'Hola desde b.cpp, Ana', 60000)
    must(fs.existsSync(path.join(folder, 'CMakeLists.txt')), 'no CMakeLists.txt was created')
    await L.shot(page, `${lang}-cmake-06-old-folder`)
    return 'a.cpp + b.cpp run; CMakeLists.txt created'
  })

  return list
}
