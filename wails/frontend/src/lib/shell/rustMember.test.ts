import { cleanup, fireEvent, render, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles } from '../bridge/languageProfiles'
import type { RunApi } from '../bridge/types'
import { setupI18n } from '../i18n'
import {
  activePath,
  connectSettings,
  fileTree,
  lastRunConfiguration,
  memberChoice,
  membersIn,
  openDialog,
  packageTask,
  profiles,
  resetRun,
  runActiveFile,
  rustAdvice
} from '../stores'
import type { RunMemberApi } from '../stores/runMember'
import RunMemberDialog from './RunMemberDialog.svelte'

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

describe('Workspace member chooser', () => {
  const asksMembers = (members: string): void => {
    vi.spyOn(bridge.run, 'run').mockRejectedValue(new Error(`run.chooseMember: ${members}`))
  }

  it('reads the member names of the error and ignores other errors', () => {
    expect(membersIn(new Error('run.chooseMember: app, cli'))).toEqual(['app', 'cli'])
    expect(membersIn('boom: run.chooseMember:solo')).toEqual(['solo'])
    expect(membersIn(new Error('tool "rustc": tool not found'))).toBeNull()
    expect(membersIn(new Error('run.chooseMember: '))).toBeNull()
  })

  it('opens the dialog instead of failing, and runs the chosen member with the same arguments', async () => {
    asksMembers('app, cli')
    const runMember = vi.fn().mockResolvedValue({})
    ;(bridge.run as RunApi & RunMemberApi).runMember = runMember
    activePath.set('taller/Cargo.toml')
    const view = render(RunMemberDialog)
    await runActiveFile(bridge)
    expect(get(memberChoice)?.members).toEqual(['app', 'cli'])
    const buttons = await waitFor(() => {
      const found = document.querySelectorAll('[data-member]')
      expect(found).toHaveLength(2)
      return found
    })
    expect(buttons[1]?.getAttribute('data-member')).toBe('cli')
    await fireEvent.click(buttons[1] as HTMLElement)
    await waitFor(() => expect(runMember).toHaveBeenCalledWith('taller/Cargo.toml', 'cli', []))
    expect(get(memberChoice)).toBeNull()
    view.unmount()
  })

  it('keeps the dialog closed for a normal run', async () => {
    activePath.set('hola-rust/src/main.rs')
    await runActiveFile(bridge)
    expect(get(memberChoice)).toBeNull()
  })
})
