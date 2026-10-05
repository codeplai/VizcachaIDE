import { describe, expect, it } from 'vitest'
import type { RunConfiguration } from '../domain'
import { isPackageCommand } from './assistant'

const run = (target: string): RunConfiguration =>
  ({
    target,
    codeLanguage: 'python',
    workingDir: '',
    mode: 'file',
    programArgs: [],
    echo: false
  }) as unknown as RunConfiguration

describe('package commands are not checked after they finish', () => {
  it('recognises go, pip and cargo commands', () => {
    for (const target of [
      'go mod tidy',
      'python -m pip --disable-pip-version-check install numpy-financial',
      'python.exe -m pip list',
      'cargo add rand'
    ]) {
      expect(isPackageCommand(run(target)), target).toBe(true)
    }
  })

  it('keeps checking real files, even in folders named like a tool', () => {
    for (const target of [
      'C:/cursos/go/main.go',
      '/home/ana/python/hola.py',
      'D:/cargo/src/main.rs',
      'go.py'
    ]) {
      expect(isPackageCommand(run(target)), target).toBe(false)
    }
  })
})
