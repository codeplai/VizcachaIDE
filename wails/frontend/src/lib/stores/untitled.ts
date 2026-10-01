import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { activePath, buffers, openTabs } from './files'

const UNTITLED_DIR = 'untitled/'

export const HELLO_PROGRAM = `package main

import "fmt"

func main() {
    fmt.Println("Hola, Go")
}
`

export const BLANK_PROGRAM = `package main

func main() {
}
`

export const isUntitled = (path: string): boolean => path.startsWith(UNTITLED_DIR)

const nextUntitledPath = (): string => {
  const taken = new Set(Object.keys(get(buffers)))
  for (let number = 1; ; number++) {
    const path = `${UNTITLED_DIR}${number === 1 ? 'main' : `main${number}`}.go`
    if (!taken.has(path)) return path
  }
}

/** Opens a file that is not on disk yet; running it uses `RunService.RunUntitled`. */
export const openUntitled = async (bridge: Bridge, source: string): Promise<string> => {
  const path = nextUntitledPath()
  buffers.update((all) => ({ ...all, [path]: source }))
  await bridge.language.openDocument(path, source)
  openTabs.update((tabs) => [...tabs, path])
  activePath.set(path)
  return path
}
