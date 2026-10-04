import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles } from '../bridge/languageProfiles'
import { setupI18n } from '../i18n'
import { activePath, consoleHistory, profiles, resetConsole } from '../stores'
import ConsolePanel from './ConsolePanel.svelte'

beforeAll(() => setupI18n('en'))
beforeEach(async () => {
  profiles.set(languageProfiles)
  await resetConsole(bridge)
  consoleHistory.set([])
})
afterEach(() => {
  cleanup()
  activePath.set(null)
})

const input = (): HTMLTextAreaElement => screen.getByRole('textbox') as HTMLTextAreaElement

const submit = async (code: string): Promise<void> => {
  await fireEvent.input(input(), { target: { value: code } })
  await fireEvent.keyDown(input(), { key: 'Enter' })
}

describe('ConsolePanel prompt', () => {
  it('is >>> when the active language is Python', async () => {
    activePath.set('C:/work/app.py')
    render(ConsolePanel)
    expect(screen.getByText('>>>')).toBeTruthy()
    await submit('2 + 3')
    await waitFor(() => expect(screen.getAllByText('>>>').length).toBe(2))
    expect(screen.getByText('5')).toBeTruthy()
  })

  it('keeps the single chevron of Go', () => {
    activePath.set('C:/work/main.go')
    render(ConsolePanel)
    expect(screen.getByText('›')).toBeTruthy()
    expect(screen.queryByText('>>>')).toBeNull()
  })

  it('shows what the backend answers when the code asks for the keyboard', async () => {
    activePath.set('C:/work/app.py')
    render(ConsolePanel)
    await submit('input()')
    await waitFor(() => expect(screen.getByText(/can't read the keyboard/)).toBeTruthy())
  })
})
