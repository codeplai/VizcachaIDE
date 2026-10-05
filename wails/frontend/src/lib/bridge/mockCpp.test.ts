import { describe, expect, it } from 'vitest'
import { parseDevQuery } from '../devQuery'
import type { EventPayloads } from '../events'
import { cppProfile, languageProfiles } from './languageProfiles'
import { createMockBridge } from './mock'

const cpp = () => createMockBridge({ sampleLanguage: 'cpp' })

const recordRun = (bridge: ReturnType<typeof cpp>['bridge']) => {
  const out = { text: [] as string[], stderr: [] as string[], finished: null as number | null }
  bridge.on('run:output', ({ stream, text }) =>
    (stream === 'stderr' ? out.stderr : out.text).push(text)
  )
  bridge.on('run:finished', ({ exitCode }) => (out.finished = exitCode))
  return out
}

describe('C++ in the mock bridge', () => {
  it('compiles and runs the sample: a Compiling notice, then Hola, C++', async () => {
    const { bridge } = cpp()
    const out = recordRun(bridge)
    let config: EventPayloads['run:started'] | null = null
    bridge.on('run:started', (started) => (config = started))
    await bridge.run.run('hola-cpp/main.cpp', [])
    expect(out.text).toEqual(['Compiling…\n', 'Hola, C++\n'])
    expect(out.finished).toBe(0)
    expect(config).toMatchObject({ codeLanguage: 'cpp', echo: true })
  })

  it('builds without running: the compiler notice only', async () => {
    const { bridge } = cpp()
    const out = recordRun(bridge)
    await bridge.run.build('hola-cpp/main.cpp', [])
    expect(out.text).toEqual(['Compiling…\n'])
    expect(out.finished).toBe(0)
  })

  it.each([
    ['en', 'C++ doesn’t know the name'],
    ['es', 'C++ no conoce el nombre']
  ] as const)('explains CPP-UNDECLARED in %s', async (language, title) => {
    const { bridge, controls } = cpp()
    await bridge.settings.save({ ...(await bridge.settings.get()), language })
    const out = recordRun(bridge)
    let explained: EventPayloads['assistant:explained'] = []
    bridge.on('assistant:explained', (items) => (explained = items))
    await controls.play('error')
    expect(explained[0]?.explanation?.explanationId).toBe('CPP-UNDECLARED')
    expect(explained[0]?.explanation?.title).toContain(title)
    expect(explained[0]?.diagnostic.location?.line).toBe(18)
    expect(out.finished).toBe(1)
    expect(out.stderr.join('')).toContain('hola-cpp/main.cpp:18:13: error:')
  })

  it.each([
    ['en', 'The program touched memory'],
    ['es', 'El programa tocó memoria']
  ] as const)('prints what it had and explains the crash in %s', async (language, title) => {
    const { bridge, controls } = cpp()
    await bridge.settings.save({ ...(await bridge.settings.get()), language })
    const out = recordRun(bridge)
    let explained: EventPayloads['assistant:explained'] = []
    bridge.on('assistant:explained', (items) => (explained = items))
    await controls.play('crash')
    expect(out.text.join('')).toContain('Hola, C++')
    expect(out.stderr).toEqual(['Segmentation fault\n'])
    expect(out.finished).toBe(139)
    expect(explained[0]?.explanation?.explanationId).toBe('CPP-SEGFAULT')
    expect(explained[0]?.explanation?.title).toContain(title)
  })

  it('debugs factorial(int n): frames, variables and the arguments of each frame', async () => {
    const { bridge, controls } = cpp()
    const stops: EventPayloads['debug:stopped'][] = []
    bridge.on('debug:stopped', (stopped) => stops.push(stopped))
    await controls.play('debug')
    const state = stops[0]
    expect(state?.frames.map((frame) => frame.function)).toEqual([
      'factorial(int)',
      'factorial(int)',
      'factorial(int)',
      'main'
    ])
    expect(state?.frames[0]?.location?.line).toBe(11)
    expect(state?.variables.map((variable) => [variable.name, variable.changed])).toEqual([
      ['n', true]
    ])
    expect((await bridge.debug.frameVariables(3)).arguments[0]?.value).toBe('3')
    expect((await bridge.debug.frameVariables(4)).locals[0]?.name).toBe('resultado')
  })

  it('serves the sample project with a .cpp and a .h file', async () => {
    const { bridge } = cpp()
    const tree = await bridge.files.listTree('')
    expect(tree.children.map((node) => node.name)).toEqual(['calculadora.h', 'main.cpp'])
    expect(await bridge.files.readFile('hola-cpp/main.cpp')).toContain('int factorial(int n)')
  })
})

describe('C++ tools and the dev bar', () => {
  it('declares the seven tools of the backend profile', () => {
    expect(cppProfile.tools.map((spec) => [spec.id, spec.role])).toEqual([
      ['cxx', 'compiler'],
      ['lldb-dap', 'debugAdapter'],
      ['clangd', 'languageServer'],
      ['clang-format', 'formatter'],
      ['cmake', 'runtime'],
      ['ninja', 'runtime'],
      ['vcpkg', 'runtime']
    ])
    expect(languageProfiles).toContain(cppProfile)
    expect(cppProfile.capabilities).toMatchObject({ build: true, console: false, debugInput: true })
  })

  it('reports each C++ tool with a version', async () => {
    const { bridge } = createMockBridge()
    const found = (await bridge.codeLanguages.tools()).filter((tool) => tool.codeLanguage === 'cpp')
    expect(found.map((tool) => tool.id)).toEqual([
      'cxx',
      'lldb-dap',
      'clangd',
      'clang-format',
      'cmake',
      'ninja',
      'vcpkg'
    ])
    expect(found.every((tool) => tool.version !== '')).toBe(true)
  })

  it('opens the C++ sample with ?language=cpp and can ask for the crash', () => {
    expect(parseDevQuery('?language=cpp&scenario=crash')).toMatchObject({
      codeLanguage: 'cpp',
      scenario: 'crash'
    })
  })
})
