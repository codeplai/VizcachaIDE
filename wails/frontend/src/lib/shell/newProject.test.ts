import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { get } from 'svelte/store'
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { bridge } from '../bridge'
import { languageProfiles } from '../bridge/languageProfiles'
import { setupI18n } from '../i18n'
import {
  activePath,
  connectSettings,
  fileTree,
  newProjectError,
  newProjectLanguage,
  newProjectLocation,
  newProjectName,
  openDialog,
  openTabs,
  profiles,
  projectNameProblem,
  startNewProject
} from '../stores'
import NewProjectDialog from './NewProjectDialog.svelte'

beforeAll(() => setupI18n('en'))

let stopSettings: (() => void) | undefined

beforeEach(async () => {
  stopSettings = await connectSettings(bridge)
  profiles.set(languageProfiles)
  await bridge.settings.save({
    ...(await bridge.settings.get()),
    enabledCodeLanguages: [],
    defaultCodeLanguage: 'go'
  })
})

afterEach(() => {
  stopSettings?.()
  cleanup()
  vi.restoreAllMocks()
  openDialog.set(null)
  fileTree.set(null)
  openTabs.set([])
  activePath.set(null)
})

const openDialogWithForm = async (): Promise<void> => {
  render(NewProjectDialog)
  await startNewProject(bridge)
  await screen.findByRole('dialog')
}

const createButton = (): HTMLButtonElement =>
  screen.getByRole('button', { name: 'Create' }) as HTMLButtonElement

const typeName = (name: string): Promise<boolean> =>
  fireEvent.input(screen.getByLabelText('Project name'), { target: { value: name } })

describe('project name rules', () => {
  it('accepts names with spaces and accents and rejects what a folder cannot have', () => {
    expect(projectNameProblem('Mi Programa Ñandú')).toBeNull()
    expect(projectNameProblem('  ')).toBe('project.errorNameEmpty')
    for (const bad of [
      'a/b',
      'a\\b',
      'a:b',
      'a*b',
      'a?b',
      'a"b',
      'a<b',
      'a>b',
      'a|b',
      '...',
      ' . '
    ]) {
      expect(projectNameProblem(bad), bad).toBe('project.errorNameInvalid')
    }
  })
})

describe('New project dialog', () => {
  it('opens empty, with Create disabled until there is a name and a location', async () => {
    await openDialogWithForm()
    expect(createButton().disabled).toBe(true)
    await typeName('hola')
    expect(createButton().disabled).toBe(true)
    await fireEvent.click(screen.getByRole('button', { name: 'Choose…' }))
    await waitFor(() => expect(get(newProjectLocation)).not.toBe(''))
    await waitFor(() => expect(createButton().disabled).toBe(false))
  })

  it('shows an invalid name as soon as it is typed and disables Create', async () => {
    await openDialogWithForm()
    newProjectLocation.set('hola-go')
    await typeName('a/b')
    expect((await screen.findByRole('alert')).textContent).toContain('cannot have')
    expect(createButton().disabled).toBe(true)
  })

  it('lists only the enabled languages and describes what will be created', async () => {
    await bridge.settings.save({
      ...(await bridge.settings.get()),
      enabledCodeLanguages: ['python', 'rust'],
      defaultCodeLanguage: 'python'
    })
    await openDialogWithForm()
    const options = [...screen.getByLabelText('Programming language').querySelectorAll('option')]
    expect(options.map((option) => option.value)).toEqual(['python', 'rust'])
    expect(get(newProjectLanguage)).toBe('python')
    expect(screen.getByText('Creates main.py')).toBeTruthy()
    await fireEvent.change(screen.getByLabelText('Programming language'), {
      target: { value: 'rust' }
    })
    expect(screen.getByText(/Cargo\.toml \+ src\/main\.rs/)).toBeTruthy()
  })

  it('creates the project, opens its folder and its main file, and closes', async () => {
    const calls: string[] = []
    const create = vi.spyOn(bridge.projects, 'create')
    const listTree = vi.spyOn(bridge.files, 'listTree')
    create.mockImplementation(async () => {
      calls.push('create')
      return { root: 'hola-go/Mi Proyecto', mainFile: 'hola-go/Mi Proyecto/main.go' }
    })
    listTree.mockImplementation(async (root) => {
      calls.push(`listTree ${root}`)
      return { name: 'Mi Proyecto', path: root, isDir: true, children: [] }
    })
    await openDialogWithForm()
    newProjectLocation.set('hola-go')
    await typeName(' Mi Proyecto ')
    await fireEvent.click(createButton())
    await waitFor(() => expect(get(activePath)).toBe('hola-go/Mi Proyecto/main.go'))
    expect(create).toHaveBeenCalledWith('go', 'hola-go', 'Mi Proyecto')
    expect(calls).toEqual(['create', 'listTree hola-go/Mi Proyecto'])
    expect(get(fileTree)?.path).toBe('hola-go/Mi Proyecto')
    expect(get(openTabs)).toContain('hola-go/Mi Proyecto/main.go')
    expect(get(openDialog)).toBeNull()
  })

  it('Enter in the name field submits the form', async () => {
    const create = vi.spyOn(bridge.projects, 'create')
    await openDialogWithForm()
    newProjectLocation.set('hola-go')
    await typeName('enter')
    await fireEvent.submit(screen.getByLabelText('Project name').closest('form') as HTMLFormElement)
    await waitFor(() => expect(create).toHaveBeenCalledWith('go', 'hola-go', 'enter'))
  })

  it('keeps the dialog and shows the translated error when the folder exists', async () => {
    vi.spyOn(bridge.projects, 'create').mockRejectedValue(
      new Error('project.errorExists: "hola-go/x"')
    )
    await openDialogWithForm()
    newProjectLocation.set('hola-go')
    await typeName('x')
    await fireEvent.click(createButton())
    expect((await screen.findByRole('alert')).textContent).toContain('already exists')
    expect(get(openDialog)).toBe('newProject')
    expect(get(newProjectName)).toBe('x')
  })

  it('shows the backend text for an unexpected failure', async () => {
    vi.spyOn(bridge.projects, 'create').mockRejectedValue(new Error('disk is full'))
    await openDialogWithForm()
    newProjectLocation.set('hola-go')
    await typeName('x')
    await fireEvent.click(createButton())
    expect((await screen.findByRole('alert')).textContent).toContain('disk is full')
    expect(get(newProjectError)?.key).toBe('project.errorFailed')
  })

  it('keeps the location when the folder dialog is cancelled', async () => {
    vi.spyOn(bridge.files, 'chooseFolder').mockResolvedValue('')
    await openDialogWithForm()
    newProjectLocation.set('antes')
    await fireEvent.click(screen.getByRole('button', { name: 'Choose…' }))
    await waitFor(() => expect(bridge.files.chooseFolder).toHaveBeenCalled())
    expect(get(newProjectLocation)).toBe('antes')
  })
})
