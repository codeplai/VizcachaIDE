// The integrated terminal (Output panel > Terminal): a real shell in the open folder with the IDE's
// toolchains first in PATH, so go, python, clang++ and cargo answer exactly as with F5. The steps
// type into xterm like a person (keyboard into its hidden textarea) and read the answers from its
// rows (`.xterm-rows`, the DOM renderer), then end the shell with Kill terminal.
import * as L from './lib.mjs'
import { must, pageOf } from './interactions.mjs'
import * as U from './ui.mjs'

const KILL = { en: 'Kill terminal', es: 'Cerrar terminal' }

// What each command prints, as a pattern its typed echo cannot match.
const COMMANDS = [
  ['go version', /go version go\d+\.\d+/],
  ['python --version', /Python 3\.\d+\.\d+/],
  // Its first line only: clang prints five, and the small panel scrolls the first one away.
  ['(clang++ --version)[0]', /clang version \d+\.\d+/],
  ['cargo --version', /cargo \d+\.\d+\.\d+/]
]

export const terminalSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const rows = () => page.evaluate(() => document.querySelector('.xterm-rows')?.innerText ?? '')
  const waitRows = (pattern, timeout) =>
    page
      .waitForFunction((source) => new RegExp(source).test(document.querySelector('.xterm-rows')?.innerText ?? ''), { timeout }, pattern.source)
      .catch(async () => {
        throw new Error(`${pattern} not in the terminal rows: ${JSON.stringify((await rows()).slice(-400))}`)
      })

  add('terminal-open', 'The Terminal tab starts a shell in the open folder (one session)', async () => {
    await U.reload(page, ctx.url)
    await page.evaluate(() => [...document.querySelectorAll('[role=tab]')].find((tab) => tab.textContent.trim() === 'Terminal').click())
    await page.waitForSelector('.xterm-rows', { timeout: 15000 })
    await waitRows(/\S/, 60000) // the prompt
    const sessions = await page.$$eval('.terminal .session', (els) => els.map((e) => e.textContent.trim()))
    must(sessions.length === 1 && sessions[0].startsWith('1:'), `sessions = ${sessions}`)
    await L.shot(page, `${lang}-terminal-01-open`)
    return `session "${sessions[0]}", rows: ${JSON.stringify((await rows()).trim().slice(0, 80))}`
  })

  for (const [command, pattern] of COMMANDS) {
    add(`terminal-${command.split(' ')[0].replace(/\W+/g, '')}`, `\`${command}\` answers in the terminal`, async () => {
      await page.focus('.xterm-helper-textarea')
      await page.keyboard.type(command)
      await page.keyboard.press('Enter')
      await waitRows(pattern, 60000)
      await L.shot(page, `${lang}-terminal-${command.split(' ')[0].replace(/\W+/g, '')}`)
      return (await rows()).split('\n').find((line) => pattern.test(line))?.trim()
    })
  }

  add('terminal-kill', 'Kill terminal ends the session and removes the view', async () => {
    await page.evaluate((label) => [...document.querySelectorAll('.terminal button')].find((b) => b.textContent.trim() === label).click(), KILL[lang])
    await page.waitForFunction(() => !document.querySelector('.terminal .session') && !document.querySelector('.xterm'), { timeout: 15000 })
    await L.shot(page, `${lang}-terminal-05-killed`)
    return 'no sessions, no xterm'
  })

  return list
}
