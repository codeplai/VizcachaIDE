import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { bridge } from '../bridge'
import { setupI18n } from '../i18n'
import { consoleHistory, resetConsole } from '../stores'
import ConsolePanel from './ConsolePanel.svelte'

beforeAll(() => setupI18n('en'))
beforeEach(async () => {
  await resetConsole(bridge)
  consoleHistory.set([])
})
afterEach(cleanup)

const input = (): HTMLTextAreaElement => screen.getByRole('textbox') as HTMLTextAreaElement

describe('ConsolePanel', () => {
  it('shows the empty state and the interpreter note', () => {
    render(ConsolePanel)
    expect(screen.getByText(/Try 2 \+ 3 or x := 10/)).toBeTruthy()
    expect(screen.getByText(/yaegi/)).toBeTruthy()
  })

  it('runs the code on Enter and shows the result', async () => {
    render(ConsolePanel)
    await fireEvent.input(input(), { target: { value: '2 + 3' } })
    await fireEvent.keyDown(input(), { key: 'Enter' })
    await waitFor(() => expect(screen.getByText('5')).toBeTruthy())
    expect(input().value).toBe('')
  })

  it('does not run on Shift+Enter', async () => {
    render(ConsolePanel)
    await fireEvent.input(input(), { target: { value: '2 + 3' } })
    await fireEvent.keyDown(input(), { key: 'Enter', shiftKey: true })
    expect(input().value).toBe('2 + 3')
    expect(screen.queryByText('5')).toBeNull()
  })

  it('recalls the previous snippet with the up arrow', async () => {
    render(ConsolePanel)
    await fireEvent.input(input(), { target: { value: '1 + 1' } })
    await fireEvent.keyDown(input(), { key: 'Enter' })
    await waitFor(() => expect(input().value).toBe(''))
    await fireEvent.keyDown(input(), { key: 'ArrowUp' })
    expect(input().value).toBe('1 + 1')
  })

  it('empties the scrollback with Reset', async () => {
    render(ConsolePanel)
    await fireEvent.input(input(), { target: { value: '7' } })
    await fireEvent.keyDown(input(), { key: 'Enter' })
    await waitFor(() => expect(screen.getAllByText('7').length).toBeGreaterThan(0))
    await fireEvent.click(screen.getByRole('button', { name: 'Reset' }))
    await waitFor(() => expect(screen.getByText(/Try 2 \+ 3/)).toBeTruthy())
  })
})
