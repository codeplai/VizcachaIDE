import { cleanup, fireEvent, render, screen } from '@testing-library/svelte'
import { tick } from 'svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles } from '../bridge/languageProfiles'
import type { PackageInfo } from '../domain'
import { setupI18n } from '../i18n'
import {
  activePath,
  fileTree,
  lastRunConfiguration,
  openDialog,
  packageTask,
  profiles,
  resetRun,
  SEARCH_DELAY_MS
} from '../stores'
import PackagesDialog from './PackagesDialog.svelte'

const found = (name: string, version = '1.0'): PackageInfo => ({
  name,
  version,
  description: `About ${name}`,
  url: ''
})

beforeAll(() => setupI18n('en'))

beforeEach(() => {
  vi.useFakeTimers()
  resetRun()
  profiles.set(languageProfiles)
  packageTask.set(null)
  lastRunConfiguration.set(null)
  fileTree.set(null)
  activePath.set('C:/work/app/main.py')
  openDialog.set('packages')
})

afterEach(() => {
  cleanup()
  vi.useRealTimers()
  vi.restoreAllMocks()
  openDialog.set(null)
  activePath.set(null)
})

const type = async (text: string) => {
  const input = await screen.findByLabelText('Package to add')
  await fireEvent.input(input, { target: { value: text } })
  return input as HTMLInputElement
}

const settle = async (ms = SEARCH_DELAY_MS) => {
  await vi.advanceTimersByTimeAsync(ms)
  await tick()
}

describe('searching a package by name', () => {
  it('waits for a quiet moment, then lists name, version and description', async () => {
    const search = vi.spyOn(bridge.packages, 'search').mockResolvedValue([found('numpy', '2.1.0')])
    render(PackagesDialog)
    await type('nu')
    await settle(SEARCH_DELAY_MS - 50)
    expect(search).not.toHaveBeenCalled()
    await settle(60)
    expect(search).toHaveBeenCalledExactlyOnceWith('python', 'nu')
    const option = await screen.findByRole('option', { name: /numpy/ })
    expect(option.textContent).toContain('2.1.0')
    expect(option.textContent).toContain('About numpy')
    expect(screen.getByRole('listbox', { name: 'Matching packages' })).toBeTruthy()
  })

  it('does not search for fewer than two characters', async () => {
    const search = vi.spyOn(bridge.packages, 'search')
    render(PackagesDialog)
    await type('n')
    await settle(1000)
    expect(search).not.toHaveBeenCalled()
  })

  it('searches only the name of a requirement', async () => {
    const search = vi.spyOn(bridge.packages, 'search').mockResolvedValue([])
    render(PackagesDialog)
    await type('numpy==2.0')
    await settle()
    expect(search).toHaveBeenCalledWith('python', 'numpy')
  })

  it('shows searching, then no results', async () => {
    let answer: (list: PackageInfo[]) => void = () => {}
    vi.spyOn(bridge.packages, 'search').mockReturnValue(
      new Promise((resolve) => (answer = resolve))
    )
    render(PackagesDialog)
    await type('zzz')
    await settle()
    expect(screen.getByText('Searching…')).toBeTruthy()
    answer([])
    await settle(0)
    expect(screen.getByText(/No packages match “zzz”/)).toBeTruthy()
  })

  it('says the index could not be searched and still lets the exact name be installed', async () => {
    vi.spyOn(bridge.packages, 'search').mockRejectedValue(new Error('down'))
    const add = vi.spyOn(bridge.packages, 'add')
    render(PackagesDialog)
    await type('requests')
    await settle()
    expect(screen.getByText(/Couldn't search the package index/)).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Install a package' }))
    expect(add).toHaveBeenCalledWith('python', 'C:/work/app', 'requests')
  })

  it('ignores an old answer when a newer search was made', async () => {
    const answers: Array<(list: PackageInfo[]) => void> = []
    vi.spyOn(bridge.packages, 'search').mockImplementation(
      () => new Promise((resolve) => answers.push(resolve))
    )
    render(PackagesDialog)
    await type('nu')
    await settle()
    await type('num')
    await settle()
    expect(answers).toHaveLength(2)
    answers[1]?.([found('numpy')])
    await settle(0)
    answers[0]?.([found('nu-old')])
    await settle(0)
    expect(screen.queryByRole('option', { name: /nu-old/ })).toBeNull()
    expect(screen.getByRole('option', { name: /numpy/ })).toBeTruthy()
  })
})

describe('choosing a result', () => {
  beforeEach(() => {
    vi.spyOn(bridge.packages, 'search').mockResolvedValue([
      found('numpy', '2.1.0'),
      found('numpy-financial')
    ])
  })

  it('fills the field, shows it as chosen and installs exactly that name', async () => {
    const add = vi.spyOn(bridge.packages, 'add')
    render(PackagesDialog)
    const input = await type('nump')
    await settle()
    await fireEvent.click(await screen.findByRole('option', { name: /numpy-financial/ }))
    expect(input.value).toBe('numpy-financial')
    expect(screen.queryByRole('listbox')).toBeNull()
    expect(screen.getByText('Chosen: numpy-financial 1.0')).toBeTruthy()
    await fireEvent.click(screen.getByRole('button', { name: 'Install a package' }))
    expect(add).toHaveBeenCalledExactlyOnceWith('python', 'C:/work/app', 'numpy-financial')
  })

  it('does not search again for the chosen name, but forgets it when the text changes', async () => {
    render(PackagesDialog)
    const input = await type('nump')
    await settle()
    await fireEvent.click(await screen.findByRole('option', { name: /numpy-financial/ }))
    await settle(1000)
    expect(bridge.packages.search).toHaveBeenCalledTimes(1)
    await fireEvent.input(input, { target: { value: 'numpy-fin' } })
    expect(screen.queryByText(/Chosen:/)).toBeNull()
  })

  it('works with the keyboard: arrows move, Enter chooses without submitting', async () => {
    const add = vi.spyOn(bridge.packages, 'add')
    render(PackagesDialog)
    const input = await type('nump')
    await settle()
    expect(input.getAttribute('aria-expanded')).toBe('true')
    await fireEvent.keyDown(input, { key: 'ArrowDown' })
    await fireEvent.keyDown(input, { key: 'ArrowDown' })
    const second = screen.getByRole('option', { name: /numpy-financial/ })
    expect(second.getAttribute('aria-selected')).toBe('true')
    expect(input.getAttribute('aria-activedescendant')).toBe(second.id)
    await fireEvent.keyDown(input, { key: 'Enter' })
    expect(input.value).toBe('numpy-financial')
    expect(add).not.toHaveBeenCalled()
  })

  it('lets Enter on a focused option choose it', async () => {
    render(PackagesDialog)
    const input = await type('nump')
    await settle()
    await fireEvent.keyDown(await screen.findByRole('option', { name: /numpy 2/ }), {
      key: 'Enter'
    })
    expect(input.value).toBe('numpy')
  })
})
