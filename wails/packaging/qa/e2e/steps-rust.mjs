// Rust checks of M3 (docs/PLAN_RUST.md section 10): compile and run with read_line, an explained
// borrow error, an explained panic, live problems and inlay hints from rust-analyzer in a Cargo
// project, Build from the More menu, the debugger (variables and keyboard input while debugging)
// and rustfmt on save. qa.mjs gives `wails dev` the development rustup (CARGO_HOME, RUSTUP_HOME)
// and Settings.toolPaths["lldb-dap"] (docs/PLAN_RUST.md section 3.3).
import fs from 'node:fs'
import path from 'node:path'
import * as L from './lib.mjs'
import * as U from './ui.mjs'
import { TEXTS } from './texts.mjs'
import { answer, must, pageOf, press, toggleBreakpoint } from './interactions.mjs'

const MOVED = { en: 'You use "nombre" after giving it away', es: 'Usas "nombre" después de haberlo entregado' }
const PANIC = { en: 'Position 5 does not exist: there are only 3 elements', es: 'La posición 5 no existe: solo hay 3 elementos' }
const BUILD = { en: 'Build', es: 'Compilar' }

const PROGRAMS = {
  'hola/main.rs': 'use std::io;\n\nfn main() {\n    println!("Nombre: ");\n    let mut nombre = String::new();\n    io::stdin().read_line(&mut nombre).unwrap();\n    println!("Hola, {}", nombre.trim());\n}\n',
  'error/main.rs': 'fn main() {\n    let nombre = String::from("Ana");\n    let otro = nombre;\n    println!("{} {}", nombre, otro);\n}\n',
  'panico/main.rs': 'fn main() {\n    let v = vec![1, 2, 3];\n    let i = v.len() + 2;\n    println!("{}", v[i]);\n}\n',
  'factorial/main.rs': 'fn factorial(n: u64) -> u64 {\n    let mut resultado: u64 = 1;\n    for i in 2..=n {\n        resultado *= i;\n    }\n    resultado\n}\n\nfn main() {\n    println!("{}", factorial(5));\n}\n',
  'entrada/main.rs': 'use std::io;\n\nfn main() {\n    println!("Color: ");\n    let mut color = String::new();\n    io::stdin().read_line(&mut color).unwrap();\n    println!("Te gusta el {}", color.trim());\n}\n',
  'formato/main.rs': 'fn main(){\nlet x=1;\n  println!("{}",x);}\n',
  'proyecto/Cargo.toml': '[package]\nname = "proyecto"\nversion = "0.1.0"\nedition = "2024"\n',
  'proyecto/src/main.rs': 'fn main() {\n    let total: i32 = "cinco";\n    println!("{}", total);\n}\n',
  'pistas/Cargo.toml': '[package]\nname = "pistas"\nversion = "0.1.0"\nedition = "2024"\n',
  'pistas/src/main.rs': 'fn main() {\n    let v = vec![1, 2, 3];\n    println!("{:?}", v);\n}\n'
}

export const rustSteps = (ctx, lang) => {
  const page = pageOf(ctx)
  const T = TEXTS[lang]
  const folder = path.join(L.PROJECT, 'qa', 'rust')
  const list = []
  const add = (id, title, run) => list.push({ id, title, run })
  const open = async (program) => {
    await L.go(page, 'FilesService', 'ListTree', path.join(folder, program.split('/')[0]))
    await U.reload(page, ctx.url)
    await U.openByName(page, path.basename(program))
    await page.waitForSelector('.cm-content', { timeout: 10000 })
  }
  const runEnabled = () =>
    page.waitForFunction((t) => [...document.querySelectorAll('.titlebar .btn')].some((b) => b.textContent.includes(t)), { timeout: 60000 }, T.run)
  const card = (text, timeout = 90000) =>
    page.waitForFunction((t) => document.querySelector('.guide')?.innerText.includes(t), { timeout }, text)

  add('rust-run', 'hola/main.rs: F5 compiles with rustc and runs it in a terminal; read_line reads Output', async () => {
    fs.rmSync(folder, { recursive: true, force: true })
    for (const [name, text] of Object.entries(PROGRAMS)) {
      fs.mkdirSync(path.join(folder, path.dirname(name)), { recursive: true })
      fs.writeFileSync(path.join(folder, name), text)
    }
    await open('hola/main.rs')
    await U.run(page)
    await U.waitOutput(page, 'Nombre:', 120000)
    await answer(page, 'Ana')
    await U.waitOutput(page, 'Hola, Ana')
    await U.waitOutput(page, T.finished)
    await L.shot(page, `${lang}-rust-01-run`)
    return 'Nombre: Ana -> Hola, Ana'
  })

  add('rust-error', 'error/main.rs: the moved value (E0382) is explained in the IDE language', async () => {
    await open('error/main.rs')
    await U.run(page)
    await U.waitOutput(page, 'E0382', 120000)
    await card(MOVED[lang])
    await L.shot(page, `${lang}-rust-02-error`)
    return `card "${MOVED[lang]}"`
  })

  add('rust-panic', 'panico/main.rs: the index out of bounds panic is explained at the student\'s line', async () => {
    await open('panico/main.rs')
    await U.run(page)
    await U.waitOutput(page, 'panicked at', 120000)
    await card(PANIC[lang])
    await L.shot(page, `${lang}-rust-03-panic`)
    return `card "${PANIC[lang]}"`
  })

  add('rust-live', 'proyecto/src/main.rs: rust-analyzer underlines the mismatched type in a Cargo project', async () => {
    await open('proyecto/main.rs')
    await page.waitForFunction(
      () => [...document.querySelectorAll('.cm-lintRange-error')].some((e) => e.textContent.includes('cinco')),
      { timeout: 150000 }
    )
    await L.shot(page, `${lang}-rust-04-live`)
    return 'underline on "cinco"'
  })

  add('rust-inlay', 'pistas/src/main.rs: an inlay hint shows the inferred type Vec<i32>', async () => {
    await open('pistas/main.rs')
    try {
      await page.waitForFunction(
        () => [...document.querySelectorAll('.cm-inlay-hint')].some((e) => e.textContent.includes('Vec<i32>')),
        { timeout: 150000 }
      )
    } catch (error) {
      const file = path.join(folder, 'pistas', 'src', 'main.rs')
      const debug = await page.evaluate(async (file) => {
        const range = { start: { file, line: 1, column: 1 }, end: { file, line: 5, column: 1 } }
        let direct
        try { direct = await window.go.bridge.LanguageService.InlayHints(range) } catch (e) { direct = 'ERR ' + String(e) }
        return { direct, hints: document.querySelectorAll('.cm-inlay-hint').length, status: document.querySelector('.statusbar, footer')?.textContent }
      }, file)
      throw new Error(`${error.message} DEBUG ${JSON.stringify(debug)}`)
    }
    await L.shot(page, `${lang}-rust-05-inlay`)
    return 'hint "Vec<i32>"'
  })

  add('rust-build', 'hola/main.rs: More > Build leaves main.exe next to the source without running it', async () => {
    const exe = path.join(folder, 'hola', process.platform === 'win32' ? 'main.exe' : 'main')
    fs.rmSync(exe, { force: true })
    await open('hola/main.rs')
    await page.click('.more-trigger')
    await press(page, '.menu-item', BUILD[lang])
    for (let i = 0; i < 120 && !fs.existsSync(exe); i++) await L.sleep(1000)
    must(fs.existsSync(exe), `${exe} was not built`)
    await runEnabled()
    must(!(await U.outputText(page)).includes('Nombre:'), 'Build ran the program')
    return path.basename(exe)
  })

  add('rust-debug', 'factorial/main.rs: breakpoint on line 4, Debug, variables n/resultado/i, Stop debugging', async () => {
    await open('factorial/main.rs')
    await toggleBreakpoint(page, 4)
    await press(page, '.titlebar .btn', T.debug)
    await page.waitForFunction((t) => document.querySelector('.guide-h .chip')?.textContent.includes(t), { timeout: 150000 }, `${T.paused} 4`)
    await page.waitForSelector('.guide .var', { timeout: 30000 })
    const vars = await page.$$eval('.guide .var', (els) => els.map((e) => [e.querySelector('.n')?.textContent.trim(), e.querySelector('.v')?.textContent.trim()]))
    const map = Object.fromEntries(vars)
    must(map.n === '5' && map.resultado === '1' && map.i === '2', `variables = ${JSON.stringify(map)}`)
    await L.shot(page, `${lang}-rust-06-debug`)
    await press(page, '.titlebar .btn.stop', T.stopDebugging)
    await runEnabled()
    return `${JSON.stringify(map)}; stopped`
  })

  add('rust-debug-input', 'entrada/main.rs: while debugging, read_line reads what is typed in Output', async () => {
    await open('entrada/main.rs')
    await press(page, '.titlebar .btn', T.debug)
    await U.waitOutput(page, 'Color:', 150000)
    await answer(page, 'verde')
    await U.waitOutput(page, 'Te gusta el verde', 30000)
    await L.shot(page, `${lang}-rust-07-debug-input`)
    await runEnabled()
    return 'Color: verde -> Te gusta el verde'
  })

  add('rust-format', 'formato/main.rs: Ctrl+S formats it with rustfmt (4 spaces)', async () => {
    const file = path.join(folder, 'formato', 'main.rs')
    fs.writeFileSync(file, PROGRAMS['formato/main.rs'])
    await open('formato/main.rs')
    await page.evaluate(() => document.querySelector('.cm-content').focus())
    await page.keyboard.down('Control')
    await page.keyboard.press('s')
    await page.keyboard.up('Control')
    let saved = ''
    for (let i = 0; i < 40 && !saved.includes('    let x = 1;'); i++) {
      await L.sleep(500)
      saved = fs.readFileSync(file, 'utf8')
    }
    must(saved.includes('    let x = 1;') && saved.includes('    println!("{}", x);'), `not formatted: ${JSON.stringify(saved)}`)
    return 'formatted with 4 spaces'
  })

  return list
}
