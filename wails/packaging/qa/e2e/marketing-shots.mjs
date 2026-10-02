// Screenshots of the new features for the website and social posts, taken from `npm run dev`
// (the mock backend, so the content is stable). Start vite first:
//   cd wails/frontend && npx vite --port 5199 --strictPort
//   node marketing-shots.mjs <output-folder>
import fs from 'node:fs'
import path from 'node:path'
import puppeteer from 'puppeteer-core'

const OUT = process.argv[2]
const BASE = 'http://localhost:5199/'
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))
fs.mkdirSync(OUT, { recursive: true })

const browser = await puppeteer.launch({
  executablePath: 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  headless: 'new'
})

const open = async (query) => {
  const page = await browser.newPage()
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 })
  await page.goto(`${BASE}?${query}`, { waitUntil: 'networkidle0' })
  await sleep(1500)
  // The mock's scenario bar is for development only: keep it out of the pictures.
  await page.addStyleTag({ content: 'div.dev { display: none !important; }' })
  await sleep(300)
  return page
}
const clickText = async (page, selector, text) => {
  const handle = await page.evaluateHandle(
    (sel, txt) => [...document.querySelectorAll(sel)].find((el) => el.textContent.trim().startsWith(txt) && el.offsetParent !== null),
    selector,
    text
  )
  await handle.asElement()?.click()
}
const shot = async (page, name) => {
  await page.screenshot({ path: path.join(OUT, `${name}.png`) })
  console.log(name)
}

for (const lang of ['es', 'en']) {
  const T = lang === 'es'
    ? { file: 'Archivo', calls: 'Llamadas', console: 'Consola' }
    : { file: 'File', calls: 'Calls', console: 'Console' }

  let page = await open(`lang=${lang}&scenario=write`)
  await page.keyboard.press('F5')
  await sleep(2500)
  await shot(page, `run.${lang}`)
  await page.close()

  page = await open(`lang=${lang}&scenario=error`)
  await shot(page, `assistant.${lang}`)
  await page.close()

  page = await open(`lang=${lang}&scenario=debug`)
  await clickText(page, '[role=tab]', T.calls)
  await sleep(1200)
  await shot(page, `calls.${lang}`)
  await page.close()

  page = await open(`lang=${lang}&scenario=write`)
  await clickText(page, '[role=tab]', T.console)
  await sleep(400)
  for (const line of ['x := 21', 'x * 2', 'x + 8']) {
    await page.type('.console textarea', line)
    await page.keyboard.press('Enter')
    await sleep(500)
  }
  await shot(page, `console.${lang}`)
  await page.close()

  page = await open(`lang=${lang}&scenario=write`)
  await clickText(page, 'button', T.file)
  await sleep(700)
  await shot(page, `file-menu.${lang}`)
  await page.close()

  page = await open(`lang=${lang}&scenario=write&theme=dark`)
  await page.keyboard.press('F5')
  await sleep(2500)
  await shot(page, `dark.${lang}`)
  await page.close()
}
await browser.close()
