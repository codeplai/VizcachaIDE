import { describe, expect, it } from 'vitest'
import { splitOutputLinks } from './outputLinks'

const places = (text: string) =>
  splitOutputLinks(text).flatMap((segment) => (segment.location ? [segment.location] : []))

describe('output links of C++', () => {
  it('links a Windows path with backslashes, a drive letter, spaces and accents', () => {
    const file = 'C:\\Users\\Ana\\mis programas ñandú\\main.cpp'
    const text = `${file}:4:5: error: 'x' was not declared in this scope`
    expect(places(text)).toEqual([{ file, line: 4, column: 5 }])
    const [link, rest] = splitOutputLinks(text)
    expect(link?.text).toBe(`${file}:4:5`)
    expect(rest?.text).toBe(": error: 'x' was not declared in this scope")
  })

  it('links relative and forward-slash paths and the headers', () => {
    expect(places('main.cpp:12:3: warning: unused variable')).toEqual([
      { file: 'main.cpp', line: 12, column: 3 }
    ])
    expect(places('In file included from src/util.hpp:7:')).toEqual([
      { file: 'src/util.hpp', line: 7, column: 1 }
    ])
    expect(places('D:/code/uno.h:2:1: error: x')).toEqual([
      { file: 'D:/code/uno.h', line: 2, column: 1 }
    ])
  })

  it("leaves GCC's linker form as plain text", () => {
    const text = "main.cpp:(.text+0x1a): undefined reference to `foo()'"
    expect(places(text)).toEqual([])
    expect(splitOutputLinks(text)).toEqual([{ text }])
    expect(places('C:\\a b\\main.o:main.cpp:(.text+0x1a): undefined reference')).toEqual([])
  })

  it('still links a linker line that does carry a line number', () => {
    expect(places("main.cpp:9: undefined reference to `foo()'")).toEqual([
      { file: 'main.cpp', line: 9, column: 1 }
    ])
  })

  it('keeps linking Go places exactly as before', () => {
    expect(places('./main.go:5:2: declared and not used')).toEqual([
      { file: './main.go', line: 5, column: 2 }
    ])
    expect(places('C:\\proj\\main.go:3')).toEqual([
      { file: 'C:\\proj\\main.go', line: 3, column: 1 }
    ])
    expect(places('File "main.py", line 3')).toEqual([])
  })
})
