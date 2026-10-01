// Keeps `wails dev` and Edge alive between runs, for exploring by hand:
//   node session.mjs start    (backs up settings, copies examples/, starts both, prints the URL)
//   node session.mjs stop     (kills only the processes it started, restores settings.json)
// Other scripts attach with `attach()` from this file.
import fs from 'node:fs'
import path from 'node:path'
import puppeteer from 'puppeteer-core'
import * as L from './lib.mjs'

const STATE = path.join(L.WQ, 'session.json')
const EDGE_PORT = 9333

const start = async () => {
  L.backupSettings()
  L.prepareProject()
  L.writeSettings({ lastFolder: L.PROJECT, language: 'en' })
  const dev = await L.startDev()
  const edge = await L.launchEdge(EDGE_PORT)
  fs.writeFileSync(STATE, JSON.stringify({ url: dev.url, devPid: dev.pid, edgePid: edge.pid, edgePort: EDGE_PORT }))
  console.log('ready', dev.url)
  // The processes must outlive this script: detach them from the exit hook.
  process.removeAllListeners('exit')
  await edge.browser.disconnect()
  process.exit(0)
}

const stop = () => {
  if (fs.existsSync(STATE)) {
    const s = JSON.parse(fs.readFileSync(STATE, 'utf8'))
    for (const pid of [s.edgePid, s.devPid]) L.forceKill(pid)
    fs.rmSync(STATE)
  }
  L.killLeftovers()
  L.restoreSettings()
  console.log('stopped')
}

export const attach = async () => {
  const s = JSON.parse(fs.readFileSync(STATE, 'utf8'))
  const browser = await puppeteer.connect({ browserURL: `http://127.0.0.1:${s.edgePort}`, defaultViewport: { width: 1280, height: 800 } })
  // Edge may open its own pages (extensions), so the app gets a tab of its own.
  const existing = (await browser.pages()).find((p) => p.url().startsWith(s.url))
  const page = existing ?? (await browser.newPage())
  return { browser, url: s.url, page }
}

if (process.argv[1] && import.meta.url.endsWith(path.basename(process.argv[1]))) {
  const cmd = process.argv[2]
  if (cmd === 'start') await start()
  else if (cmd === 'stop') stop()
  else console.log('usage: node session.mjs start|stop')
}
