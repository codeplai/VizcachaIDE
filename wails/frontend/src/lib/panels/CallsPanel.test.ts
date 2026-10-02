import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { bridge, mockControls, type Unsubscribe } from '../bridge'
import { setupI18n } from '../i18n'
import { connectStores, revealRequest } from '../stores'
import CallsPanel from './CallsPanel.svelte'

let stop: Unsubscribe | undefined

beforeAll(() => {
  setupI18n('en')
})

beforeEach(async () => {
  stop = await connectStores(bridge)
})

afterEach(() => {
  cleanup()
  stop?.()
})

describe('CallsPanel', () => {
  it('says there is nothing to show when the program is not paused', () => {
    render(CallsPanel)
    expect(screen.getByText(/the program is not paused/)).toBeTruthy()
  })

  it('draws the recursion as nested boxes with their arguments', async () => {
    render(CallsPanel)
    await mockControls?.play('debug')

    await waitFor(() => expect(screen.getByText('factorial(n=1)')).toBeTruthy())
    expect(screen.getByText('factorial(n=2)')).toBeTruthy()
    expect(screen.getByText('factorial(n=3)')).toBeTruthy()
    const boxes = screen.getAllByTestId('call-box')
    expect(boxes).toHaveLength(5) // main, 3 x factorial, sumar
    expect(boxes[0]?.textContent).toContain('main()')
    expect(boxes[0]?.contains(boxes[1] as Node)).toBe(true) // nested, not side by side
    expect(boxes.at(-1)?.className).toContain('current')
    expect(screen.getByText('sumar(a=1, b=0)')).toBeTruthy()
  })

  it('takes the editor to the line of the box that was clicked', async () => {
    render(CallsPanel)
    await mockControls?.play('debug')
    const button = await screen.findByRole('button', { name: 'Go to line 11 of main' })

    await fireEvent.click(button)

    await waitFor(() => expect(get(revealRequest)?.location.line).toBe(11))
  })
})
