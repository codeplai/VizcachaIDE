import { describe, expect, it } from 'vitest'
import { splitOutputLinks } from './outputLinks'

const places = (text: string) =>
  splitOutputLinks(text).flatMap((segment) => (segment.location ? [segment.location] : []))

describe('output links of Rust', () => {
  it("links rustc's arrow line with a relative Windows path", () => {
    const text = '  --> src\\main.rs:5:10'
    expect(places(text)).toEqual([{ file: 'src\\main.rs', line: 5, column: 10 }])
    expect(splitOutputLinks(text)[0]?.text).toBe('  --> ')
  })

  it("links the panic's place and drops its closing colon", () => {
    const text = "thread 'main' (35236) panicked at src\\main.rs:4:21:"
    expect(places(text)).toEqual([{ file: 'src\\main.rs', line: 4, column: 21 }])
    expect(splitOutputLinks(text).at(-1)?.text).toBe(':')
  })

  it("links the backtrace's frame of the student's code", () => {
    expect(places('             at .\\src\\main.rs:4:21')).toEqual([
      { file: '.\\src\\main.rs', line: 4, column: 21 }
    ])
    expect(places('             at ./src/bin/otro.rs:9:5')).toEqual([
      { file: './src/bin/otro.rs', line: 9, column: 5 }
    ])
  })

  it('links a Windows path with spaces and accents', () => {
    const file = 'C:\\Users\\María\\mis programas\\hola rust\\src\\main.rs'
    const text = `  --> ${file}:12:3`
    expect(places(text)).toEqual([{ file, line: 12, column: 3 }])
  })

  it.each([
    '             at /rustc/17067e9ac6d7e98f1c1a3a2b6d9e6a3a5c2f1b00/library/core/src/panicking.rs:221:5',
    '   3: core::panicking::panic_bounds_check\n             at /rustc/abcdef0123456789/library/core/src/panicking.rs:276:5',
    '             at C:\\Users\\ana\\.toolchain\\lib\\rustlib\\src\\rust\\library\\std\\src\\rt.rs:206:18'
  ])('does not link the frames of the standard library: %s', (text) => {
    expect(places(text)).toEqual([])
    expect(splitOutputLinks(text)).toEqual([{ text }])
  })

  it('links the student frame and leaves the library one in the same output', () => {
    const text = 'at /rustc/abc1234/library/core/src/option.rs:1:1 then at .\\src\\main.rs:3:7'
    expect(places(text)).toEqual([{ file: '.\\src\\main.rs', line: 3, column: 7 }])
  })
})
