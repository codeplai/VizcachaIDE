import { describe, expect, it } from 'vitest'
import { selectBridge } from '.'
import type { EventName, EventPayloads } from '../events'
import { Events } from '../events'
import { createMockBridge } from './mock'

const record = () => {
  const mock = createMockBridge()
  const seen: EventName[] = []
  const names = [
    'run:started',
    'run:output',
    'run:finished',
    'debug:stopped',
    'debug:terminated',
    'lsp:diagnostics',
    'assistant:explained'
  ] as const
  names.forEach((name) => mock.bridge.on(name, () => seen.push(name)))
  return { ...mock, seen }
}

describe('mock bridge', () => {
  it('is selected when there is no Wails backend', () => {
    const selection = selectBridge()
    expect(selection.bridge.isMock).toBe(true)
    expect(selection.mockControls).not.toBeNull()
  })

  it('runs successfully in the write scenario', async () => {
    const { bridge, seen } = record()
    await bridge.run.run('hola-go/main.go', [])
    expect(seen).toContain('run:started')
    expect(seen).toContain('run:output')
    expect(seen[seen.length - 1]).toBe('run:finished')
  })

  it('publishes an explained error in the error scenario', async () => {
    const { bridge, controls } = createMockBridge()
    let explained: EventPayloads['assistant:explained'] = []
    bridge.on('assistant:explained', (items) => (explained = items))
    await controls.play('error')
    expect(controls.scenario()).toBe('error')
    expect(explained).toHaveLength(1)
    expect(explained[0]?.explanation?.explanationId).toBe('E-UNUSED-VAR')
  })

  it('stops at line 6 in the debug scenario and ends when stopped by the user', async () => {
    const { bridge, controls } = createMockBridge()
    let line = 0
    let exitCode = 0
    bridge.on('debug:stopped', (state) => (line = state.frames[0]?.location?.line ?? 0))
    bridge.on('debug:terminated', (payload) => (exitCode = payload.exitCode))
    await controls.play('debug')
    expect(line).toBe(6)
    await bridge.debug.stop()
    expect(exitCode).toBe(-1)
  })

  it('stops listening after unsubscribing', async () => {
    const { bridge } = createMockBridge()
    let count = 0
    const off = bridge.on('run:finished', () => count++)
    await bridge.run.run('a.go', [])
    off()
    await bridge.run.run('a.go', [])
    expect(count).toBe(1)
  })

  it('notifies settings changes', async () => {
    const { bridge } = createMockBridge()
    let language = ''
    bridge.on('settings:changed', (settings) => (language = settings.language))
    const current = await bridge.settings.get()
    await bridge.settings.save({ ...current, language: 'es' })
    expect(language).toBe('es')
  })

  it('emits every event of the contract in some scenario', async () => {
    const { bridge, controls } = createMockBridge()
    const seen = new Set<string>()
    Object.values(Events).forEach((name) => bridge.on(name, () => seen.add(name)))
    await controls.play('write')
    await controls.play('error')
    await controls.play('debug')
    await bridge.debug.requestVariables(11)
    await bridge.language.openDocument('a.go', '')
    await bridge.settings.save(await bridge.settings.get())
    await bridge.updates.check()
    await new Promise((resolve) => setTimeout(resolve, 400))
    await bridge.debug.stop()
    expect([...seen].sort()).toEqual(Object.values(Events).sort())
  })
})
