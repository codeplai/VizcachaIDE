// Tests that need real Windows windows, not only the page:
// 1. the compiled app (build/bin/vizcacha.exe, from `wails build`) remembers its size, position
//    and maximized state, and never opens off-screen;
// 2. the native Open and Save as dialogs: the page under `wails dev` presses the buttons, the
//    real Go backend opens the Windows dialog, and uia-dialog.ps1 answers it (no keystrokes).
//   node native-app.mjs   → results in %TEMP%/wq/native-app.json
// Real windows open on the desktop. Only processes started here are closed (by PID);
// settings.json and window.json are backed up and restored.
import { spawn, spawnSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'

const EXE = path.join(L.WAILS_DIR, 'build', 'bin', 'vizcacha.exe')
const WINDOW_FILE = path.join(path.dirname(L.SETTINGS_FILE), 'window.json')
const WINDOW_BACKUP = path.join(L.WQ, 'window-backup.json')
const results = []
const must = (condition, message) => {
  if (!condition) throw new Error(message)
}
const step = async (id, run) => {
  try {
    results.push({ id, ok: true, detail: String(await run()) })
  } catch (error) {
    results.push({ id, ok: false, detail: String(error?.message ?? error) })
  }
  const last = results[results.length - 1]
  console.log(`${last.ok ? 'PASS' : 'FAIL'} ${id}: ${last.detail.slice(0, 260)}`)
}

const WIN32 = `
Add-Type @"
using System; using System.Runtime.InteropServices;
public static class W {
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
  [DllImport("user32.dll")] public static extern bool MoveWindow(IntPtr h, int x, int y, int w, int hh, bool r);
  [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr h, int c);
  [DllImport("user32.dll")] public static extern bool IsZoomed(IntPtr h);
  [DllImport("user32.dll")] public static extern bool PostMessage(IntPtr h, uint m, IntPtr w, IntPtr l);
  public struct RECT { public int Left, Top, Right, Bottom; }
}
"@
`
const ps = (script) => spawnSync('powershell.exe', ['-NoProfile', '-Command', WIN32 + script], { encoding: 'utf8' }).stdout.trim()
const mainWindow = (pid) => `$h = (Get-Process -Id ${pid}).MainWindowHandle;`
const rectOf = (pid) =>
  JSON.parse(ps(`${mainWindow(pid)} $r = New-Object W+RECT; [W]::GetWindowRect($h, [ref]$r) | Out-Null; @{x=$r.Left;y=$r.Top;w=$r.Right-$r.Left;h=$r.Bottom-$r.Top;max=[W]::IsZoomed($h)} | ConvertTo-Json -Compress`))

const launch = async () => {
  const child = spawn(EXE, [], { stdio: 'ignore' })
  for (let i = 0; i < 60; i++) {
    await L.sleep(500)
    const handle = ps(`(Get-Process -Id ${child.pid} -ErrorAction SilentlyContinue).MainWindowHandle`)
    if (handle && handle !== '0') break
  }
  await L.sleep(2500)
  return child
}
/** Closes the window like the user does (WM_CLOSE), so OnBeforeClose saves the state. */
const closeWindow = async (child) => {
  ps(`${mainWindow(child.pid)} [W]::PostMessage($h, 0x10, [IntPtr]::Zero, [IntPtr]::Zero) | Out-Null`)
  for (let i = 0; i < 40 && child.exitCode === null; i++) await L.sleep(250)
  if (child.exitCode === null) L.forceKill(child.pid)
}
/** Starts answering the dialog in the background; resolves with the script's output. */
const answerDialog = (title, filePath) =>
  new Promise((resolve) => {
    const script = path.join(L.HERE, 'uia-dialog.ps1')
    const child = spawn('powershell.exe', ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', script, '-Title', title, '-FilePath', filePath])
    let output = ''
    child.stdout.on('data', (data) => (output += data))
    child.stderr.on('data', (data) => (output += data))
    child.on('exit', () => resolve(output.trim().split(/\r?\n/)[0]))
  })

L.killLeftovers()
L.backupSettings()
if (fs.existsSync(WINDOW_FILE)) fs.copyFileSync(WINDOW_FILE, WINDOW_BACKUP)
const project = L.prepareProject()
L.writeSettings({ lastFolder: project, firstRun: false, language: 'es' })
fs.rmSync(WINDOW_FILE, { force: true })
let app = null
let dev = null
let edge = null
try {
  await step('window-remembers-size-and-position', async () => {
    app = await launch()
    ps(`${mainWindow(app.pid)} [W]::ShowWindow($h, 9) | Out-Null; [W]::MoveWindow($h, 140, 90, 1150, 720, $true) | Out-Null`)
    await L.sleep(800)
    const before = rectOf(app.pid)
    await closeWindow(app)
    app = await launch()
    const after = rectOf(app.pid)
    await closeWindow(app)
    const close = (a, b) => Math.abs(a - b) <= 16
    must(close(after.w, before.w) && close(after.h, before.h) && close(after.x, before.x) && close(after.y, before.y), `${JSON.stringify(before)} → ${JSON.stringify(after)}`)
    return `before ${JSON.stringify(before)} after ${JSON.stringify(after)}`
  })

  await step('window-remembers-maximized', async () => {
    app = await launch()
    ps(`${mainWindow(app.pid)} [W]::ShowWindow($h, 3) | Out-Null`)
    await L.sleep(800)
    await closeWindow(app)
    app = await launch()
    const after = rectOf(app.pid)
    await closeWindow(app)
    must(after.max === true, `not maximized after restart: ${JSON.stringify(after)}`)
    return 'maximized again after restart'
  })

  await step('window-off-screen-is-centered', async () => {
    fs.writeFileSync(WINDOW_FILE, JSON.stringify({ width: 1100, height: 700, x: 9000, y: 9000, maximised: false, hasPosition: true }))
    app = await launch()
    const rect = rectOf(app.pid)
    await closeWindow(app)
    must(rect.x < 3000 && rect.y < 3000 && rect.x > -200, `opened off-screen: ${JSON.stringify(rect)}`)
    return `saved at 9000,9000 → opened at ${rect.x},${rect.y}`
  })

  dev = await L.startDev()
  edge = await L.launchEdge()
  const page = await L.openApp(edge.browser, dev.url)
  await L.sleep(2500)
  const titlebar = () => page.evaluate(() => document.querySelector('.titlebar')?.innerText ?? '')

  await step('native-open-dialog', async () => {
    const target = path.join(project, 'leer', 'leer.go')
    const answered = answerDialog('Abrir archivo', target)
    await page.click('button[aria-label^="Abrir archivo"]')
    const sent = await answered
    await page.waitForFunction(() => document.querySelector('.titlebar')?.innerText.includes('leer.go'), { timeout: 15000 }).catch(() => {})
    must((await titlebar()).includes('leer.go'), `${sent}; title bar: ${await titlebar()}`)
    return `${sent}; tab leer.go opened`
  })

  await step('native-save-as-dialog', async () => {
    const target = path.join(project, 'guardado_qa.go')
    await page.click('button[aria-label^="Nuevo archivo"]')
    await L.sleep(800)
    const answered = answerDialog('Guardar archivo', target)
    await page.click('button[aria-label^="Guardar"]')
    const sent = await answered
    for (let i = 0; i < 40 && !fs.existsSync(target); i++) await L.sleep(250)
    must(fs.existsSync(target), `${sent}; the file was not written`)
    await page.waitForFunction(() => document.querySelector('.titlebar')?.innerText.includes('guardado_qa.go'), { timeout: 10000 }).catch(() => {})
    await L.shot(page, 'native-save-as')
    return `${sent}; ${fs.statSync(target).size} bytes written; title bar: ${(await titlebar()).split('\n')[0]}`
  })
} finally {
  if (app && app.exitCode === null) await closeWindow(app)
  if (edge) await edge.stop()
  if (dev) dev.stop()
  L.restoreSettings()
  if (fs.existsSync(WINDOW_BACKUP)) fs.copyFileSync(WINDOW_BACKUP, WINDOW_FILE)
  else fs.rmSync(WINDOW_FILE, { force: true })
  fs.rmSync(WINDOW_BACKUP, { force: true })
}
fs.writeFileSync(`${L.WQ}/native-app.json`, JSON.stringify(results, null, 2))
console.log(`\n${results.filter((r) => r.ok).length}/${results.length} steps passed`)
process.exit(results.every((r) => r.ok) ? 0 : 1)
