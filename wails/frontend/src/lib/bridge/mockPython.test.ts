import { describe, expect, it } from 'vitest'
import { parseDevQuery } from '../devQuery'
import type { EventPayloads } from '../events'
import { createMockBridge } from './mock'
import { languageProfiles, pythonProfile } from './languageProfiles'

const python = () => createMockBridge({ sampleLanguage: 'python' })

describe('Python in the mock bridge', () => {
  it('runs the sample and prints Hola, Python', async () => {
    const { bridge } = python()
    const output: string[] = []
    let config: EventPayloads['run:started'] | null = null
    bridge.on('run:output', ({ text }) => output.push(text))
    bridge.on('run:started', (started) => (config = started))
    await bridge.run.run('hola-py/main.py', [])
    expect(output).toEqual(['Hola, Python\n'])
    expect(config).toMatchObject({ codeLanguage: 'python', echo: true })
  })

  it('still refuses C++, which has no adapter yet', async () => {
    const { bridge } = python()
    await expect(bridge.run.run('main.cpp', [])).rejects.toThrow(/does not support/)
  })

  it.each([
    ['en', 'Python doesn’t know the name'],
    ['es', 'Python no conoce el nombre']
  ] as const)('explains the NameError in %s', async (language, title) => {
    const { bridge, controls } = python()
    await bridge.settings.save({ ...(await bridge.settings.get()), language })
    let explained: EventPayloads['assistant:explained'] = []
    bridge.on('assistant:explained', (items) => (explained = items))
    await controls.play('error')
    expect(explained[0]?.explanation?.explanationId).toBe('E-PY-NAME-ERROR')
    expect(explained[0]?.explanation?.title).toContain(title)
    expect(explained[0]?.diagnostic.message).toContain('not defined')
  })

  it('debugs factorial(n): frames, variables and the arguments of each frame', async () => {
    const { bridge, controls } = python()
    const stops: EventPayloads['debug:stopped'][] = []
    bridge.on('debug:stopped', (stopped) => stops.push(stopped))
    await controls.play('debug')
    const state = stops[0]
    expect(state?.frames.map((frame) => frame.function)).toEqual([
      'factorial',
      'factorial',
      'factorial',
      '<module>'
    ])
    expect(state?.frames[0]?.location?.line).toBe(8)
    expect(state?.variables.map((variable) => variable.name)).toEqual(['n'])
    expect((await bridge.debug.frameVariables(3)).arguments[0]?.value).toBe('3')
  })

  it('serves the sample project with .py files', async () => {
    const { bridge } = python()
    const tree = await bridge.files.listTree('')
    expect(tree.children.map((node) => node.name)).toEqual(['main.py', 'calculadora.py'])
    expect(await bridge.files.readFile('hola-py/main.py')).toContain('def factorial(n):')
  })

  it('evaluates in the console and says it cannot read the keyboard', async () => {
    const { bridge } = python()
    expect((await bridge.console.eval('python', '2 + 3')).result).toBe('5')
    expect((await bridge.console.eval('python', 'print("hi")')).output).toBe('hi\n')
    expect((await bridge.console.eval('python', 'input()')).error).toMatch(
      /can't read the keyboard/
    )
  })
})

describe('Python tools and the dev bar', () => {
  it('declares the four tools of the backend profile, three of them provided by python', () => {
    expect(pythonProfile.tools.map((spec) => [spec.id, spec.providedBy])).toEqual([
      ['python', ''],
      ['debugpy', 'python'],
      ['pylsp', 'python'],
      ['ruff', 'python']
    ])
    expect(languageProfiles).toContain(pythonProfile)
    expect(pythonProfile.capabilities.debugInput).toBe(true)
  })

  it('reports each Python tool with a version', async () => {
    const { bridge } = createMockBridge()
    const found = (await bridge.codeLanguages.tools()).filter(
      (tool) => tool.codeLanguage === 'python'
    )
    expect(found.map((tool) => tool.id)).toEqual(['python', 'debugpy', 'pylsp', 'ruff'])
    expect(found.every((tool) => tool.version !== '')).toBe(true)
  })

  it('opens the Python sample with ?language=python, apart from ?lang', () => {
    expect(parseDevQuery('?lang=es&language=python')).toMatchObject({
      language: 'es',
      codeLanguage: 'python'
    })
    expect(parseDevQuery('').codeLanguage).toBeNull()
  })
})
