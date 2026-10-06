// Shared helpers for the E2E scripts: `wails dev` with the real Go backend, Edge headless over CDP.
// Only the processes started here are ever killed (by PID, with their children).
import { spawn, spawnSync } from 'node:child_process'
import fs from 'node:fs'
import http from 'node:http'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import puppeteer from 'puppeteer-core'

export const HERE = path.dirname(fileURLToPath(import.meta.url))
export const WAILS_DIR = path.resolve(HERE, '../../..')
export const REPO = path.resolve(WAILS_DIR, '..')
export const EDGE = process.env.EDGE_PATH || 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'
export const WQ = path.join(os.tmpdir(), 'wq')
/** A Python with debugpy, pylsp, pyflakes and ruff for the Python steps (docs/PLAN_PYTHON.md section 0). */
export const DEV_PYTHON =
  process.env.VIZCACHA_TEST_PYTHON ||
  path.join(WAILS_DIR, '.venv-py312', process.platform === 'win32' ? 'Scripts/python.exe' : 'bin/python')
/** A clang++ with lldb-dap, clangd and clang-format next to it for the C++ steps (docs/PLAN_CPP.md section 0). */
export const DEV_CXX = path.join(
  process.env.VIZCACHA_TEST_LLVM_BIN || path.join(WAILS_DIR, '.toolchain-dev', 'llvm-mingw-20260922-ucrt-x86_64', 'bin'),
  process.platform === 'win32' ? 'clang++.exe' : 'clang++'
)
/** CMake, Ninja and vcpkg of .toolchain-dev for the C++ steps (docs/PLAN_CPP_CMAKE.md section 8). */
const DEV_TOOLCHAIN = path.join(WAILS_DIR, '.toolchain-dev')
export const DEV_CMAKE_BIN = process.env.VIZCACHA_TEST_CMAKE_BIN || path.join(DEV_TOOLCHAIN, 'cmake', 'bin')
export const DEV_VCPKG_ROOT = process.env.VIZCACHA_TEST_VCPKG_ROOT || path.join(DEV_TOOLCHAIN, 'vcpkg')
export const DEV_VCPKG_SEED = process.env.VIZCACHA_TEST_VCPKG_SEED || path.join(DEV_TOOLCHAIN, 'vcpkg-seed')
const exe = (name) => (process.platform === 'win32' ? `${name}.exe` : name)
/** Settings.toolPaths for every C++ project: the compiler, CMake, Ninja and vcpkg's root. */
export const DEV_CPP_TOOLS = {
  cxx: DEV_CXX,
  cmake: path.join(DEV_CMAKE_BIN, exe('cmake')),
  ninja: path.join(DEV_CMAKE_BIN, exe('ninja')),
  vcpkg: DEV_VCPKG_ROOT
}
/** The seed of vcpkg's own downloads is found through this variable, which `wails dev` inherits. */
export const useDevVcpkgSeed = () => { process.env.VIZCACHA_TEST_VCPKG_SEED = DEV_VCPKG_SEED }
/** The development rustup (docs/PLAN_RUST.md section 3.3): CARGO_HOME and RUSTUP_HOME for wails dev. */
export const DEV_RUST = {
  CARGO_HOME: process.env.VIZCACHA_TEST_CARGO_HOME || path.join(WAILS_DIR, '.toolchain-dev', 'cargo'),
  RUSTUP_HOME: process.env.VIZCACHA_TEST_RUSTUP_HOME || path.join(WAILS_DIR, '.toolchain-dev', 'rustup')
}
/** lldb-dap for the Rust steps: chosen in Settings, never on PATH (it would shadow rustc's MinGW linker). */
export const DEV_LLDB_DAP = path.join(path.dirname(DEV_CXX), process.platform === 'win32' ? 'lldb-dap.exe' : 'lldb-dap')
export const PROJECT = path.join(WQ, 'proj')
export const SETTINGS_DIR = path.join(process.env.APPDATA || path.join(os.homedir(), '.config'), 'VizcachaIDE')
export const SETTINGS_FILE = path.join(SETTINGS_DIR, 'settings.json')
const BACKUP = path.join(WQ, 'settings-backup.json')
const BACKUP_MARK = path.join(WQ, 'settings-backup.state')

// Wails' default (34115) falls in a Windows excluded port range on some PCs (bind: access denied).
export const DEV_PORT = process.env.WAILS_DEV_PORT || '35115'
export const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const pids = new Set()

const processTable = () => {
  const ps = spawnSync(
    'powershell.exe',
    ['-NoProfile', '-Command', 'Get-CimInstance Win32_Process | Select-Object ProcessId,ParentProcessId,ExecutablePath | ConvertTo-Json -Compress'],
    { encoding: 'utf8', maxBuffer: 64 * 1024 * 1024 }
  )
  try { return JSON.parse(ps.stdout) } catch { return [] }
}

/** pid plus every descendant, read from the process table before anything is killed. */
const withDescendants = (root, table) => {
  const found = new Set([root])
  for (let grew = true; grew;) {
    grew = false
    for (const p of table) {
      if (found.has(p.ParentProcessId) && !found.has(p.ProcessId)) { found.add(p.ProcessId); grew = true }
    }
  }
  return [...found].reverse()
}

/** Kills a process and all its descendants by PID. Only call it with PIDs this tooling started. */
export const forceKill = (pid) => {
  if (!pid) return
  for (const p of withDescendants(pid, processTable())) spawnSync('taskkill', ['/PID', String(p), '/F'], { stdio: 'ignore' })
}

/** Kills what a dead session left behind: binaries that live under this worktree's wails/ folder. */
export const killLeftovers = () => {
  const root = WAILS_DIR.toLowerCase()
  for (const p of processTable()) {
    if (p.ExecutablePath && p.ExecutablePath.toLowerCase().startsWith(root)) forceKill(p.ProcessId)
  }
}

export const killTree = (pid) => {
  if (!pid || !pids.has(pid)) return
  pids.delete(pid)
  forceKill(pid)
}
export const killAll = () => [...pids].forEach(killTree)
process.on('exit', killAll)
for (const sig of ['SIGINT', 'SIGTERM']) process.on(sig, () => { killAll(); process.exit(130) })

/** Extra programs for the checks the examples do not cover; each lives in its own folder (one main per package). */
export const QA_FILES = {
  'qa/args/main.go': 'package main\n\nimport (\n\t"fmt"\n\t"os"\n\t"strings"\n)\n\nfunc main() {\n\tfmt.Println("args=[" + strings.Join(os.Args[1:], "|") + "]")\n}\n',
  'qa/loop/main.go': 'package main\n\nimport (\n\t"fmt"\n\t"time"\n)\n\nfunc main() {\n\tfor i := 1; ; i++ {\n\t\tfmt.Println("tick", i)\n\t\ttime.Sleep(200 * time.Millisecond)\n\t}\n}\n',
  'qa/fmt/main.go': 'package main\nimport "fmt"\nfunc main(){\nfmt.Println( "formatted" )\n      x:=1\n_ = x\n}\n',
  'qa/complete/main.go': 'package main\n\nimport "fmt"\n\nfunc add(a, b int) int { return a + b }\n\nfunc main() {\n\tfmt.Println(add(1, 2))\n}\n',
  'qa/mod/go.mod': 'module example.com/qa\n\ngo 1.21\n',
  'qa/mod/main.go': 'package main\n\nimport "fmt"\n\nfunc main() { fmt.Println(greeting()) }\n',
  'qa/mod/helper.go': 'package main\n\nfunc greeting() string { return "module ok" }\n'
}

/** Copies examples/ into %TEMP%/wq/proj (a fresh copy every time). */
export const prepareProject = () => {
  fs.rmSync(PROJECT, { recursive: true, force: true })
  fs.mkdirSync(PROJECT, { recursive: true })
  fs.cpSync(path.join(REPO, 'examples'), PROJECT, { recursive: true })
  for (const [file, text] of Object.entries(QA_FILES)) {
    fs.mkdirSync(path.dirname(path.join(PROJECT, file)), { recursive: true })
    fs.writeFileSync(path.join(PROJECT, file), text)
  }
  return PROJECT
}

/** Backs up the user's real settings.json once (if there is one). Restore with restoreSettings(). */
export const backupSettings = () => {
  fs.mkdirSync(WQ, { recursive: true })
  if (fs.existsSync(BACKUP_MARK)) return // an earlier run already saved the original
  if (fs.existsSync(SETTINGS_FILE)) {
    fs.copyFileSync(SETTINGS_FILE, BACKUP)
    fs.writeFileSync(BACKUP_MARK, 'file')
  } else {
    fs.writeFileSync(BACKUP_MARK, 'none')
  }
}

export const restoreSettings = () => {
  if (!fs.existsSync(BACKUP_MARK)) return
  const state = fs.readFileSync(BACKUP_MARK, 'utf8')
  if (state === 'file') fs.copyFileSync(BACKUP, SETTINGS_FILE)
  else fs.rmSync(SETTINGS_FILE, { force: true })
  fs.rmSync(BACKUP_MARK, { force: true })
  fs.rmSync(BACKUP, { force: true })
}

export const writeSettings = (patch) => {
  fs.mkdirSync(SETTINGS_DIR, { recursive: true })
  const base = { language: 'auto', theme: 'system', fontSize: 14, goPath: '', delvePath: '', goplsPath: '', firstRun: false, lastFolder: '', formatOnSave: true }
  fs.writeFileSync(SETTINGS_FILE, JSON.stringify({ ...base, ...patch }, null, 2))
}
export const readSettings = () => JSON.parse(fs.readFileSync(SETTINGS_FILE, 'utf8'))

const ping = (url) =>
  new Promise((resolve) => {
    http.get(url, (res) => { res.resume(); resolve(res.statusCode === 200) }).on('error', () => resolve(false))
  })

/** Starts `wails dev` and resolves with { url, pid, stop }. */
export const startDev = async () => {
  fs.mkdirSync(WQ, { recursive: true })
  const logFile = path.join(WQ, 'wails-dev.log')
  const fd = fs.openSync(logFile, 'w')
  const child = spawn('wails.exe', ['dev', '-loglevel', 'Warning', '-devserver', `localhost:${DEV_PORT}`], {
    cwd: WAILS_DIR, windowsHide: true, stdio: ['ignore', fd, fd]
  })
  pids.add(child.pid)
  let exited = false
  child.on('exit', () => { exited = true })
  const output = () => fs.readFileSync(logFile, 'utf8').replace(/\[[0-9;]*m/g, '')
  const deadline = Date.now() + 240000
  let url = ''
  while (Date.now() < deadline && !exited) {
    const m = /DevServer URL:\s*(http:\/\/localhost:\d+)/i.exec(output())
    const devUrl = m ? m[1] : `http://localhost:${DEV_PORT}` // 5173 is Vite, not the Wails dev server
    if (await ping(devUrl + '/')) { url = devUrl; break }
    await sleep(1000)
  }
  if (!url) { killTree(child.pid); throw new Error('wails dev did not serve the dev server\n' + output()) }
  await sleep(2000)
  return { url, pid: child.pid, output, stop: () => killTree(child.pid) }
}

/** Starts Edge headless with CDP and connects puppeteer-core to it. */
export const launchEdge = async (port = 9333) => {
  const profile = path.join(WQ, 'edge-profile')
  fs.rmSync(profile, { recursive: true, force: true })
  const child = spawn(
    EDGE,
    ['--headless=new', `--remote-debugging-port=${port}`, `--user-data-dir=${profile}`, '--window-size=1280,800', '--no-first-run', '--disable-gpu', '--force-device-scale-factor=1', 'about:blank'],
    { windowsHide: true, stdio: 'ignore', detached: true }
  )
  pids.add(child.pid)
  for (let i = 0; i < 40; i++) {
    if (await ping(`http://127.0.0.1:${port}/json/version`)) break
    await sleep(500)
  }
  const browser = await puppeteer.connect({ browserURL: `http://127.0.0.1:${port}`, defaultViewport: { width: 1280, height: 800 } })
  return { browser, pid: child.pid, stop: async () => { try { await browser.disconnect() } catch { /* already gone */ } killTree(child.pid) } }
}

export const shot = async (page, name) => {
  fs.mkdirSync(WQ, { recursive: true })
  const file = path.join(WQ, `${name}.png`)
  await page.screenshot({ path: file })
  return file
}

/** Opens the app page and waits until the real backend answers (window.go exists). */
export const openApp = async (browser, url) => {
  const page = await browser.newPage()
  await page.setViewport({ width: 1280, height: 800 })
  await page.goto(url, { waitUntil: 'load' })
  await page.waitForFunction(() => !!window.go && document.querySelector('.titlebar'), { timeout: 30000 })
  return page
}

/** Clicks the first element of `selector` whose text matches `text`. */
export const clickText = async (page, selector, text) => {
  const handle = await page.evaluateHandle(
    (sel, txt) => [...document.querySelectorAll(sel)].find((el) => el.textContent.trim().includes(txt) && el.offsetParent !== null),
    selector,
    text
  )
  const el = handle.asElement()
  if (!el) throw new Error(`no ${selector} with text "${text}"`)
  await el.click()
}

export const bodyText = (page) => page.evaluate(() => document.body.innerText)

export const waitText = (page, text, timeout = 30000) =>
  page.waitForFunction((t) => document.body.innerText.includes(t), { timeout }, text)

/** Calls a bound Go method directly: svc('FilesService', 'ListTree', dir). */
export const go = (page, service, method, ...args) =>
  page.evaluate((s, m, a) => window.go.bridge[s][m](...a), service, method, args)
