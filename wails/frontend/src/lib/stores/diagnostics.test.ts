import { describe, expect, it } from 'vitest'
import type { ExplainedDiagnostic } from '../domain'
import { problemsOfFile, sameFile } from './diagnostics'

const at = (file: string): ExplainedDiagnostic => ({
  diagnostic: {
    location: { file, line: 1, column: 1 },
    end: null,
    severity: 'error',
    message: 'm',
    rawText: '',
    source: 'go',
    code: ''
  },
  explanation: null
})

describe('problems of one file', () => {
  it('matches the same file with different path styles', () => {
    expect(sameFile('proj/main.go', 'main.go')).toBe(true)
    expect(sameFile('C:\\proj\\main.go', 'C:/proj/main.go')).toBe(true)
    expect(sameFile('proj/main.go', 'other.go')).toBe(false)
  })

  it('keeps only the problems of the open file', () => {
    const items = [at('proj/main.go'), at('proj/calc.go')]
    expect(problemsOfFile(items, 'proj/main.go')).toHaveLength(1)
    expect(problemsOfFile(items, null)).toEqual([])
  })
})
