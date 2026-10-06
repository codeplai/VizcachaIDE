// Parity QA of the Wails variant against the REAL backend (Go, Delve, gopls).
//
//   cd wails/packaging/qa/e2e && npm install
//   node qa.mjs                 # EN and ES, settings persistence and first run
//   node qa.mjs --lang es       # one language only
//   node qa.mjs --only persist  # one phase: en | es | persist | firstrun | python | cpp | cpp-cmake | rust | newproject | terminal | files | multi
//
// It backs up the real settings.json, starts `wails dev` (the frontend with the real bindings is
// served on http://localhost:35115 here; Wails' default 34115 is reserved on some Windows PCs),
// drives it with Edge headless over CDP and puts screenshots in %TEMP%\wq. At the end it kills
// only what it started and restores settings.json. Exit code 1 if any step failed.
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { steps } from './steps.mjs'
import { persistenceSteps, firstRunSteps } from './settings-steps.mjs'
import { pythonSteps } from './steps-python.mjs'
import { cppSteps } from './steps-cpp.mjs'
import { cppCmakeSteps } from './steps-cpp-cmake.mjs'
import { rustSteps } from './steps-rust.mjs'
import { multifileSteps } from './steps-multifile.mjs'
import { newProjectSteps } from './steps-newproject.mjs'
import { terminalSteps } from './steps-terminal.mjs'
import { filesSteps } from './steps-files.mjs'

const args = process.argv.slice(2)
const option = (name) => (args.includes(name) ? args[args.indexOf(name) + 1] : null)
const onlyLang = option('--lang')
const only = option('--only')
const wants = (phase) => !only || only === phase

const results = []
const record = (phase, step, ok, evidence) => {
  results.push({ phase, id: step.id, title: step.title, ok, evidence })
  console.log(`${ok ? 'PASS' : 'FAIL'} [${phase}] ${step.id}: ${String(evidence).slice(0, 220).replace(/\n/g, ' ')}`)
}

/**
 * Starts the app and runs the steps that `build(ctx)` returns, in order. `ctx` has the page and the
 * URL, and `restart()` stops `wails dev` and Edge and starts them again (settings persistence).
 */
const session = async (phase, settings, build) => {
  L.writeSettings({ lastFolder: L.PROJECT, firstRun: false, ...settings })
  let dev
  let edge
  const ctx = { page: null, url: '', lang: phase }
  const start = async () => {
    dev = await L.startDev()
    edge = await L.launchEdge()
    ctx.url = dev.url
    ctx.page = await edge.browser.newPage()
    ctx.page.on('pageerror', (e) => {
      if (!/reading 'nodes'/.test(e.message)) console.log(`  [pageerror] ${(process.env.QA_STACK ? e.stack : e.message).slice(0, process.env.QA_STACK ? 1500 : 160)}`)
    })
    await U.reload(ctx.page, dev.url)
  }
  const stop = async () => {
    await edge?.stop()
    dev?.stop()
    await L.sleep(1500)
    L.killLeftovers()
  }
  ctx.restart = async () => {
    await stop()
    await start()
  }
  try {
    await start()
    for (const step of build(ctx)) {
      if (option('--step') && !option('--step').split(',').includes(step.id)) continue
      try {
        if (step.silent) {
          await step.run()
          continue
        }
        const alive = await Promise.race([ctx.page.evaluate(() => 'yes'), L.sleep(5000).then(() => 'NO')])
        if (alive !== 'yes') console.log(`  [harness] the page does not answer before ${step.id}`)
        record(phase, step, true, await step.run())
      } catch (error) {
        record(phase, step, false, error.message)
        await L.shot(ctx.page, `${phase}-FAIL-${step.id}`).catch(() => null)
      }
    }
  } finally {
    await stop()
  }
}

const main = async () => {
  fs.mkdirSync(L.WQ, { recursive: true })
  L.killLeftovers()
  L.backupSettings()
  L.useDevVcpkgSeed()
  L.prepareProject()
  try {
    for (const lang of ['en', 'es']) {
      if (onlyLang && onlyLang !== lang) continue
      if (!wants(lang)) continue
      await session(lang, { language: lang }, (ctx) => steps(ctx, lang))
    }
    if (wants('persist') && !onlyLang) await session('persist', { language: 'en' }, (ctx) => persistenceSteps(ctx))
    if (wants('firstrun') && !onlyLang) await session('firstrun', { language: 'en', firstRun: true }, (ctx) => firstRunSteps(ctx))
    // Python (M1): the development interpreter of docs/PLAN_PYTHON.md section 0, or VIZCACHA_TEST_PYTHON.
    if (wants('python')) {
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`python-${lang}`, { language: lang, toolPaths: { python: L.DEV_PYTHON } }, (ctx) => pythonSteps(ctx, lang))
      }
    }
    // C++ (M2): the development llvm-mingw of docs/PLAN_CPP.md section 0, or VIZCACHA_TEST_LLVM_BIN.
    if (wants('cpp')) {
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`cpp-${lang}`, { language: lang, toolPaths: { ...L.DEV_CPP_TOOLS } }, (ctx) => cppSteps(ctx, lang))
      }
    }
    // C++ with CMake and vcpkg (M4): a new project, fmt from the Packages dialog, F5, debug, an old folder.
    if (wants('cpp-cmake')) {
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`cpp-cmake-${lang}`, { language: lang, toolPaths: { ...L.DEV_CPP_TOOLS } }, (ctx) => cppCmakeSteps(ctx, lang))
      }
    }
    // Rust (M3): the development rustup of docs/PLAN_RUST.md section 3.3, inherited by wails dev.
    if (wants('rust')) {
      Object.assign(process.env, L.DEV_RUST)
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`rust-${lang}`, { language: lang, toolPaths: { 'lldb-dap': L.DEV_LLDB_DAP } }, (ctx) => rustSteps(ctx, lang))
      }
    }
    // File > New project in every language (after 2.4.0).
    if (wants('newproject')) {
      Object.assign(process.env, L.DEV_RUST)
      const toolPaths = { python: L.DEV_PYTHON, ...L.DEV_CPP_TOOLS, 'lldb-dap': L.DEV_LLDB_DAP }
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`newproject-${lang}`, { language: lang, toolPaths }, (ctx) => newProjectSteps(ctx, lang))
      }
    }
    // The integrated terminal: go, python, clang++ and cargo work in it as with F5.
    if (wants('terminal')) {
      Object.assign(process.env, L.DEV_RUST)
      const toolPaths = { python: L.DEV_PYTHON, ...L.DEV_CPP_TOOLS, 'lldb-dap': L.DEV_LLDB_DAP }
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`terminal-${lang}`, { language: lang, toolPaths }, (ctx) => terminalSteps(ctx, lang))
      }
    }
    // The Files panel: multi-select, copy, cut, paste, duplicate, drag and drop, delete several.
    if (wants('files')) {
      for (const lang of ['en', 'es']) {
        if (onlyLang && onlyLang !== lang) continue
        await session(`files-${lang}`, { language: lang }, (ctx) => filesSteps(ctx, lang))
      }
    }
    // Multi-file projects with an external library in every language (only on request: it
    // downloads packages and takes a while).
    if (only === 'multi') {
      Object.assign(process.env, L.DEV_RUST)
      const lang = onlyLang ?? 'en'
      const toolPaths = { python: L.DEV_PYTHON, ...L.DEV_CPP_TOOLS, 'lldb-dap': L.DEV_LLDB_DAP }
      for (const codeLanguage of ['go', 'python', 'cpp', 'rust']) {
        if (option('--code') && option('--code') !== codeLanguage) continue
        await session(`multi-${codeLanguage}`, { language: lang, toolPaths }, (ctx) => multifileSteps(ctx, lang, codeLanguage))
      }
    }
  } finally {
    L.killLeftovers()
    L.restoreSettings()
  }
  fs.writeFileSync(path.join(L.WQ, 'results.json'), JSON.stringify(results, null, 2))
  const failed = results.filter((r) => !r.ok)
  console.log(`\n${results.length - failed.length}/${results.length} steps passed; results in ${path.join(L.WQ, 'results.json')}`)
  process.exitCode = failed.length ? 1 : 0
}

await main()
