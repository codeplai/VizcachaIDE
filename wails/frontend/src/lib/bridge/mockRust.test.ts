import { describe, expect, it } from 'vitest'
import type { RunMemberApi } from '../stores/runMember'
import { parseDevQuery } from '../devQuery'
import type { EventPayloads } from '../events'
import { languageProfiles, rustProfile } from './languageProfiles'
import { createMockBridge } from './mock'
import type { RunApi } from './types'

const rust = () => createMockBridge({ sampleLanguage: 'rust' })
const MAIN = 'hola-rust/src/main.rs'

const recordRun = (bridge: ReturnType<typeof rust>['bridge']) => {
  const out = { text: [] as string[], stderr: [] as string[], finished: null as number | null }
  bridge.on('run:output', ({ stream, text }) =>
    (stream === 'stderr' ? out.stderr : out.text).push(text)
  )
  bridge.on('run:finished', ({ exitCode }) => (out.finished = exitCode))
  return out
}

describe('Rust in the mock bridge', () => {
  it('compiles and runs the sample: the Compiling notice, then Hola, Rust', async () => {
    const { bridge } = rust()
    const out = recordRun(bridge)
    let config: EventPayloads['run:started'] | null = null
    bridge.on('run:started', (started) => (config = started))
    await bridge.run.run(MAIN, [])
    expect(out.text[0]).toMatch(/^Compiling /)
    expect(out.text.at(-1)).toBe('Hola, Rust\n')
    expect(out.finished).toBe(0)
    expect(config).toMatchObject({ codeLanguage: 'rust', echo: true })
  })

  it('builds without running', async () => {
    const { bridge } = rust()
    const out = recordRun(bridge)
    await bridge.run.build(MAIN, [])
    expect(out.text).toHaveLength(1)
    expect(out.finished).toBe(0)
  })

  it.each([
    ['en', 'was moved'],
    ['es', 'se movió']
  ] as const)(
    'explains RS-MOVED in %s with the moved variable underlined',
    async (language, part) => {
      const { bridge, controls } = rust()
      await bridge.settings.save({ ...(await bridge.settings.get()), language })
      const out = recordRun(bridge)
      let explained: EventPayloads['assistant:explained'] = []
      bridge.on('assistant:explained', (items) => (explained = items))
      await controls.play('error')
      const item = explained[0]
      expect(item?.explanation?.explanationId).toBe('RS-MOVED')
      expect(item?.explanation?.title).toContain(part)
      expect(item?.diagnostic.code).toBe('E0382')
      expect(item?.diagnostic.location).toMatchObject({ line: 14, column: 26 })
      expect(item?.diagnostic.end).toMatchObject({ line: 14, column: 32 })
      expect(out.finished).toBe(1)
      const raw = out.stderr.join('')
      expect(raw).toContain('error[E0382]: borrow of moved value')
      expect(raw).toContain('--> src\\main.rs:14:26')
    }
  )

  it.each([
    ['en', 'a position that doesn’t exist'],
    ['es', 'una posición que no existe']
  ] as const)('prints the panic and explains RS-PANIC-INDEX in %s', async (language, part) => {
    const { bridge, controls } = rust()
    await bridge.settings.save({ ...(await bridge.settings.get()), language })
    const out = recordRun(bridge)
    let explained: EventPayloads['assistant:explained'] = []
    bridge.on('assistant:explained', (items) => (explained = items))
    await controls.play('crash')
    expect(out.text.join('')).toContain('Hola, Rust')
    expect(out.stderr.join('')).toContain('panicked at src\\main.rs:13:20:')
    expect(out.finished).toBe(101)
    expect(explained[0]?.explanation?.explanationId).toBe('RS-PANIC-INDEX')
    expect(explained[0]?.explanation?.title).toContain(part)
  })

  it('debugs factorial(n: u64): frames, variables and the arguments of each frame', async () => {
    const { bridge, controls } = rust()
    const stops: EventPayloads['debug:stopped'][] = []
    bridge.on('debug:stopped', (stopped) => stops.push(stopped))
    await controls.play('debug')
    const state = stops[0]
    expect(state?.frames.map((frame) => frame.function)).toEqual([
      'hola_rust::factorial',
      'hola_rust::factorial',
      'hola_rust::factorial',
      'hola_rust::main'
    ])
    expect(state?.frames[0]?.location?.line).toBe(5)
    expect(state?.variables.map((v) => [v.name, v.typeName, v.changed])).toEqual([
      ['n', 'u64', true]
    ])
    expect((await bridge.debug.frameVariables(3)).arguments[0]?.value).toBe('3')
    expect((await bridge.debug.frameVariables(4)).locals.map((v) => v.name)).toEqual([
      'nombre',
      'resultado'
    ])
  })

  it('serves a Cargo project with its manifest and src/main.rs', async () => {
    const { bridge } = rust()
    const tree = await bridge.files.listTree('')
    expect(tree.children.map((node) => node.name)).toEqual(['Cargo.toml', 'src'])
    expect(tree.children[1]?.children.map((node) => node.name)).toEqual(['main.rs'])
    expect(await bridge.files.readFile(MAIN)).toContain('fn factorial(n: u64) -> u64')
    expect(await bridge.files.readFile(MAIN)).toContain('read_line')
  })

  it('leaves Go, Python and C++ on their own samples', async () => {
    const { bridge } = createMockBridge()
    const out = recordRun(bridge)
    await bridge.run.run('hola-go/main.go', [])
    expect(out.text).toEqual(['Hola, Go\n'])
  })
})

describe('Rust tools, profile and the dev bar', () => {
  it('declares the six tools of the backend profile', () => {
    expect(rustProfile.tools.map((spec) => [spec.id, spec.role, spec.providedBy])).toEqual([
      ['rustc', 'compiler', ''],
      ['cargo', 'runtime', ''],
      ['rust-analyzer', 'languageServer', ''],
      ['lldb-dap', 'debugAdapter', ''],
      ['clippy', 'compiler', 'cargo'],
      ['rustfmt', 'formatter', 'cargo']
    ])
    expect(languageProfiles).toContain(rustProfile)
    expect(rustProfile.capabilities).toMatchObject({
      build: true,
      console: false,
      format: true,
      check: true,
      debugInput: true,
      packageActions: ['init', 'add', 'remove', 'list']
    })
  })

  it('reports each Rust tool with a version', async () => {
    const { bridge } = createMockBridge()
    const found = (await bridge.codeLanguages.tools()).filter(
      (tool) => tool.codeLanguage === 'rust'
    )
    expect(found.map((tool) => tool.id)).toEqual(rustProfile.tools.map((spec) => spec.id))
    expect(found.every((tool) => tool.version !== '')).toBe(true)
  })

  it('opens the Rust sample with ?language=rust and can ask for the panic', () => {
    expect(parseDevQuery('?language=rust&scenario=crash')).toMatchObject({
      codeLanguage: 'rust',
      scenario: 'crash'
    })
  })

  it('asks which member to run for a virtual workspace, then runs the chosen one', async () => {
    const { bridge } = rust()
    const run = bridge.run as RunApi & RunMemberApi
    await expect(run.run('taller-rust/Cargo.toml', [])).rejects.toThrow(
      'run.chooseMember: app, cli'
    )
    const out = recordRun(bridge)
    await run.runMember('taller-rust/Cargo.toml', 'cli', [])
    expect(out.finished).toBe(0)
  })
})
