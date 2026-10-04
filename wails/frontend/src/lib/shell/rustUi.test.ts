import { language as languageFacet } from '@codemirror/language'
import { EditorState } from '@codemirror/state'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles, rustProfile } from '../bridge/languageProfiles'
import { languageExtensionsFor } from '../editor/languageSupport'
import { newFileTemplates } from '../editor/templates'
import { setupI18n } from '../i18n'
import {
  activePath,
  connectSettings,
  fileTree,
  isValidPackage,
  isValidProjectName,
  lastRunConfiguration,
  memberChoice,
  openDialog,
  packageTask,
  profiles,
  resetRun,
  rustAdvice
} from '../stores'
import MoreMenu from './MoreMenu.svelte'
import PackagesDialog from './PackagesDialog.svelte'

beforeAll(() => setupI18n('en'))

let stopSettings: (() => void) | undefined

beforeEach(async () => {
  stopSettings = await connectSettings(bridge)
  profiles.set(languageProfiles)
  await bridge.settings.save({
    ...(await bridge.settings.get()),
    enabledCodeLanguages: ['rust'],
    defaultCodeLanguage: 'rust'
  })
})

afterEach(() => {
  stopSettings?.()
  cleanup()
  vi.restoreAllMocks()
  activePath.set(null)
  fileTree.set(null)
  openDialog.set(null)
  packageTask.set(null)
  memberChoice.set(null)
  rustAdvice.set([])
  lastRunConfiguration.set(null)
  resetRun()
})

describe('Rust in the editor', () => {
  it('has the greeting template and a blank one', () => {
    const { hello, blank, extension } = newFileTemplates.rust
    expect(hello).toContain('println!("Hola, Rust");')
    expect(blank).toContain('fn main()')
    expect(extension).toBe('.rs')
  })

  it('highlights .rs files with the Rust language and 4 spaces', () => {
    const state = EditorState.create({
      doc: 'fn main() { println!("hola"); }\n',
      extensions: languageExtensionsFor('src/main.rs', rustProfile)
    })
    expect(state.facet(languageFacet)?.name).toBe('rust')
    expect(state.tabSize).toBe(4)
  })
})

describe('Build entry of the More menu for Rust', () => {
  it('is there for a saved Rust file and builds it through the bridge', async () => {
    const build = vi.spyOn(bridge.run, 'build')
    activePath.set('hola-rust/src/main.rs')
    render(MoreMenu)
    await fireEvent.keyDown(screen.getByRole('button', { name: /More/ }), { key: 'Enter' })
    const items = await screen.findAllByRole('menuitem')
    const entry = items.find((item) => /build/i.test(item.textContent ?? ''))
    expect(entry).toBeTruthy()
    await fireEvent.click(entry as HTMLElement)
    await waitFor(() => expect(build).toHaveBeenCalledWith('hola-rust/src/main.rs', []))
  })
})

describe('Packages for Rust', () => {
  beforeEach(() => {
    activePath.set('C:/work/hola rust/src/main.rs')
    openDialog.set('packages')
  })

  const cargoFolder = {
    name: 'hola rust',
    path: 'C:/work/hola rust',
    isDir: true,
    children: [
      { name: 'Cargo.toml', path: 'C:/work/hola rust/Cargo.toml', isDir: false, children: [] }
    ]
  }

  it('declares the Cargo verbs and no tidy', () => {
    expect(rustProfile.capabilities.packageActions).toEqual(['init', 'add', 'remove', 'list'])
    expect(rustProfile.capabilities.packageActions).not.toContain('tidy')
  })

  it('offers init without a Cargo.toml, with a crate name made from the folder', async () => {
    fileTree.set({ ...cargoFolder, children: [] })
    const init = vi.spyOn(bridge.packages, 'init')
    const view = render(PackagesDialog)
    const input = (await screen.findByRole('textbox')) as HTMLInputElement
    expect(input.value).toBe('hola-rust')
    await fireEvent.click(view.baseElement.querySelector('button[type="submit"]') as HTMLElement)
    expect(init).toHaveBeenCalledWith('rust', 'C:/work/hola rust', 'hola-rust')
  })

  it('rejects a crate name that Cargo would refuse', async () => {
    fileTree.set({ ...cargoFolder, children: [] })
    const view = render(PackagesDialog)
    await fireEvent.input(await screen.findByRole('textbox'), { target: { value: '9 lives' } })
    const submit = view.baseElement.querySelector('button[type="submit"]') as HTMLButtonElement
    expect(submit.disabled).toBe(true)
  })

  it('adds, removes and lists with a Cargo.toml, and has no Tidy', async () => {
    fileTree.set(cargoFolder)
    const add = vi.spyOn(bridge.packages, 'add')
    const remove = vi.spyOn(bridge.packages, 'remove')
    const list = vi.spyOn(bridge.packages, 'list')
    const tidy = vi.spyOn(bridge.packages, 'tidy')
    const view = render(PackagesDialog)
    await fireEvent.input(await screen.findByRole('textbox'), { target: { value: 'serde@1.0' } })
    await fireEvent.click(view.baseElement.querySelector('button[type="submit"]') as HTMLElement)
    expect(add).toHaveBeenCalledWith('rust', 'C:/work/hola rust', 'serde@1.0')
    const buttons = [...view.baseElement.querySelectorAll('.dlg-button')]
    // add, remove, list (and the dialog's own Close)
    await fireEvent.click(buttons[1] as HTMLElement)
    expect(remove).toHaveBeenCalledWith('rust', 'C:/work/hola rust', 'serde@1.0')
    await fireEvent.click(buttons[2] as HTMLElement)
    expect(list).toHaveBeenCalledWith('rust', 'C:/work/hola rust')
    expect(tidy).not.toHaveBeenCalled()
  })

  it('validates crates and names per language and leaves the others as they were', () => {
    expect(isValidPackage('serde', 'rust')).toBe(true)
    expect(isValidPackage('tokio@1.40.0', 'rust')).toBe(true)
    expect(isValidPackage('two words', 'rust')).toBe(false)
    expect(isValidPackage('requests==2.0', 'python')).toBe(true)
    expect(isValidProjectName('hola_rust', 'rust')).toBe(true)
    expect(isValidProjectName('example.com/hola', 'rust')).toBe(false)
    expect(isValidProjectName('example.com/hola')).toBe(true)
  })
})
