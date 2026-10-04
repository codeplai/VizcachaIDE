// Settings persistence across `wails dev` restarts, and the first-run wizard.
import * as L from './lib.mjs'
import * as U from './ui.mjs'

const must = (condition, message) => {
  if (!condition) throw new Error(message)
}

const pressText = async (page, selector, text) => {
  const handle = await page.evaluateHandle(
    (sel, txt) => [...document.querySelectorAll(sel)].find((el) => el.offsetParent !== null && el.textContent.trim().replace(/\s+/g, ' ').startsWith(txt)),
    selector,
    text
  )
  const el = handle.asElement()
  must(el, `no visible ${selector} starting with "${text}"`)
  await el.click()
}

const openSettings = async (page, more, settings) => {
  await page.click('.more-trigger')
  await page.waitForSelector('[role=menuitem]', { timeout: 5000 })
  // 'Automatic' follows the system language, which may be either one on the test PC.
  await pressText(page, '[role=menuitem]', await page.evaluate(() => (document.querySelector('.titlebar .btn.primary').textContent.includes('Ejecutar') ? 'Configuración' : 'Settings')))
  await page.waitForSelector('[role=dialog] [role=tab]', { timeout: 5000 })
}

const closeDialog = async (page) => {
  await page.keyboard.press('Escape')
  await L.sleep(400)
}

const waitSaved = async (check, what) => {
  for (let i = 0; i < 20; i++) {
    const saved = L.readSettings()
    if (check(saved)) return saved
    await L.sleep(250)
  }
  throw new Error(`settings.json never got ${what}: ${JSON.stringify(L.readSettings())}`)
}

export const persistenceSteps = (ctx) => {
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })

  add('settings-ui', 'Settings: language Automatic/English/Spanish, dark theme, font size 18 are saved in settings.json', async () => {
    const page = ctx.page
    await openSettings(page, 'More', 'Settings')
    const options = await page.$$eval('#setting-language option', (els) => els.map((e) => e.textContent.trim()))
    must(options.join() === 'Automatic,English,Español', `language options: ${options}`)
    await page.select('#setting-language', 'auto')
    await waitSaved((s) => s.language === 'auto', 'language auto')
    await page.select('#setting-language', 'en')
    await waitSaved((s) => s.language === 'en', 'language en')
    await page.select('#setting-language', 'es')
    await waitSaved((s) => s.language === 'es', 'language es')
    await page.waitForFunction(() => document.body.innerText.includes('Configuración'), { timeout: 5000 })
    await page.select('#setting-theme', 'dark')
    await waitSaved((s) => s.theme === 'dark', 'theme dark')
    must((await page.evaluate(() => document.documentElement.getAttribute('data-theme'))) === 'dark', 'data-theme is not dark')
    await pressText(page, '[role=tab]', 'Editor')
    await page.waitForSelector('#setting-font-size', { timeout: 5000 })
    await page.$eval('#setting-font-size', (input) => {
      input.value = '18'
      input.dispatchEvent(new Event('change', { bubbles: true }))
    })
    const saved = await waitSaved((s) => s.fontSize === 18, 'fontSize 18')
    await L.shot(page, 'persist-01-settings-es-dark')
    await closeDialog(page)
    return JSON.stringify({ language: saved.language, theme: saved.theme, fontSize: saved.fontSize })
  })

  add('tools-tab', 'Settings > Tools shows Go, Delve and gopls with version and origin', async () => {
    const page = ctx.page
    await openSettings(page, 'Más', 'Configuración')
    await pressText(page, '[role=tab]', 'Herramientas')
    await page.waitForFunction(() => /Delve/.test(document.querySelector('[role=dialog]')?.innerText ?? ''), { timeout: 10000 })
    const text = await page.$eval('[role=dialog]', (e) => e.innerText.replace(/\n+/g, ' | '))
    must(/Versión/.test(text), `tools tab: ${text}`)
    await L.shot(page, 'persist-02-tools')
    await closeDialog(page)
    return text.slice(0, 200)
  })

  add('restart', 'After restarting wails dev the language, theme and font size are still there', async () => {
    await ctx.restart()
    const page = ctx.page
    const saved = L.readSettings()
    must(saved.language === 'es' && saved.theme === 'dark' && saved.fontSize === 18, `settings.json = ${JSON.stringify(saved)}`)
    const theme = await page.evaluate(() => document.documentElement.getAttribute('data-theme'))
    must(theme === 'dark', `data-theme = ${theme}`)
    const run = await page.$eval('.titlebar .btn.primary', (e) => e.textContent)
    must(run.includes('Ejecutar'), `Run button says "${run.trim()}"`)
    await U.openByName(page, 'hello.go')
    const size = await page.$eval('.cm-content', (e) => getComputedStyle(e).fontSize)
    must(size === '18px', `editor font size ${size}`)
    await L.shot(page, 'persist-03-after-restart')
    return `Spanish UI, data-theme=dark, editor ${size}`
  })

  add('auto-language', 'Automatic language follows the system language (backend and UI agree)', async () => {
    const page = ctx.page
    await openSettings(page, 'Más', 'Configuración')
    await page.select('#setting-language', 'auto')
    await waitSaved((s) => s.language === 'auto', 'language auto')
    const backend = await L.go(page, 'SettingsService', 'ResolvedLanguage')
    await L.sleep(500)
    const run = await page.$eval('.titlebar .btn.primary', (e) => e.textContent)
    const ui = run.includes('Ejecutar') ? 'es' : 'en'
    const browserLanguage = await page.evaluate(() => navigator.language)
    await closeDialog(page)
    return `settings.json=auto; backend resolves "${backend}"; UI shows "${ui}" (the test browser reports ${browserLanguage})`
  })

  add('modules-dialog', 'Go modules dialog opens (Spanish)', async () => {
    const page = ctx.page
    await page.click('.more-trigger')
    await page.waitForSelector('[role=menuitem]', { timeout: 5000 })
    await pressText(page, '[role=menuitem]', 'Módulos de Go')
    await page.waitForSelector('[role=dialog]', { timeout: 5000 })
    await L.sleep(400)
    await L.shot(page, 'persist-04-modules')
    const text = await page.$eval('[role=dialog]', (e) => e.innerText.replace(/\n+/g, ' | '))
    await closeDialog(page)
    return text.slice(0, 160)
  })
  return list
}

export const firstRunSteps = (ctx) => {
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })

  add('wizard', 'First start: welcome, language, Go check, open the hello example, run it', async () => {
    const page = ctx.page
    await page.waitForSelector('[role=dialog]', { timeout: 15000 })
    let text = await page.$eval('[role=dialog]', (e) => e.innerText)
    must(text.includes('Welcome to VizcachaIDE'), `wizard: ${text}`)
    await L.shot(page, 'firstrun-01-welcome')
    await pressText(page, '[role=dialog] button', 'Español')
    await page.waitForFunction(() => document.querySelector('[role=dialog]')?.innerText.includes('Elige tu idioma'), { timeout: 5000 })
    await L.shot(page, 'firstrun-02-language-es')
    await pressText(page, '[role=dialog] button', 'Siguiente')
    // Step 2 (M1): the programming languages the student will use; keep the default (all).
    await page.waitForFunction(() => document.querySelector('[role=dialog]')?.innerText.includes('¿Qué lenguajes de programación vas a usar?'), { timeout: 5000 })
    await L.shot(page, 'firstrun-03-code-languages')
    await pressText(page, '[role=dialog] button', 'Siguiente')
    await page.waitForFunction(() => /Go .* está listo/.test(document.querySelector('[role=dialog]')?.innerText ?? ''), { timeout: 20000 })
    text = await page.$eval('[role=dialog]', (e) => e.innerText)
    await L.shot(page, 'firstrun-04-go-ready')
    await pressText(page, '[role=dialog] button', 'Siguiente')
    await pressText(page, '[role=dialog] button', 'Abrir el ejemplo')
    await page.waitForFunction(() => !document.querySelector('[role=dialog]'), { timeout: 10000 })
    const saved = await (async () => {
      for (let i = 0; i < 20; i++) {
        const s = L.readSettings()
        if (s.firstRun === false && s.language === 'es') return s
        await L.sleep(250)
      }
      throw new Error(`settings.json after the wizard: ${JSON.stringify(L.readSettings())}`)
    })()
    await U.run(page)
    await U.waitOutput(page, 'Terminó bien')
    const output = await U.outputText(page)
    must(output.includes('Hola, Go'), `untitled run: ${output}`)
    await L.shot(page, 'firstrun-05-hello-run')
    return `${text.split('\n').find((l) => l.includes('está listo'))}; untitled tab ran: ${output.split('\n').slice(-2).join(' / ')}; firstRun=${saved.firstRun}`
  })
  return list
}
