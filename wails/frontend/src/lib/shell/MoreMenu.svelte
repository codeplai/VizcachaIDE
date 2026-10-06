<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { t } from '../i18n'
  import { bridge } from '../bridge'
  import {
    activeCodeLanguage,
    activePath,
    buildActiveFile,
    capabilities,
    checkForUpdates,
    codeLanguageOf,
    editorBridge,
    isUntitled,
    openDialog,
    settingsTab
  } from '../stores'

  // The package manager of the open file's language; languages without one have no entry.
  const hasPackages = $derived(($capabilities?.packageActions?.length ?? 0) > 0)
  // Build needs a saved file: an untitled one has no folder to put the program in.
  const canBuild = $derived(
    ($capabilities?.build ?? false) && $activePath !== null && !isUntitled($activePath)
  )
  // Rename and references need the language server of the open file.
  const canRefactor = $derived($activePath !== null && codeLanguageOf($activePath) !== null)
  const entries = $derived([
    ...(canRefactor
      ? [
          {
            label: 'refactor.menuRename',
            inEditor: true,
            run: () => editorBridge()?.renameSymbol()
          },
          {
            label: 'refactor.menuReferences',
            inEditor: true,
            run: () => editorBridge()?.findReferences()
          }
        ]
      : []),
    ...(canBuild ? [{ label: 'actions.build', run: () => void buildActiveFile(bridge) }] : []),
    ...(hasPackages
      ? [
          {
            label: $activeCodeLanguage === 'go' ? 'shell.goModules' : 'packages.title',
            run: () => openDialog.set('packages')
          }
        ]
      : []),
    { label: 'panels.settings', run: () => openDialog.set('settings') },
    {
      label: 'updates.checkMenu',
      run: () => {
        settingsTab.set('updates')
        openDialog.set('settings')
        void checkForUpdates(bridge)
      }
    },
    { label: 'shell.about', run: () => openDialog.set('about') }
  ])

  // Entries that act in the editor run once the menu has closed, so the menu does not take the
  // focus back from the editor (the rename box would lose it at once).
  let afterClose: (() => void) | null = null
  const choose = (entry: { run: () => void; inEditor?: boolean }): void => {
    if (entry.inEditor) afterClose = entry.run
    else entry.run()
  }
  const closed = (event: Event): void => {
    if (!afterClose) return
    event.preventDefault()
    const run = afterClose
    afterClose = null
    run()
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="more-trigger">
    {$t('actions.more')}
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content class="menu" align="end" sideOffset={6} onCloseAutoFocus={closed}>
      {#each entries as entry (entry.label)}
        <DropdownMenu.Item class="menu-item" onSelect={() => choose(entry)}>
          {$t(entry.label)}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.more-trigger) {
    border: 1px solid transparent;
    background: none;
    color: var(--muted);
    font: 600 14px var(--ui);
    padding: 7px 8px;
    border-radius: 8px;
    cursor: pointer;
    white-space: nowrap;
  }
  :global(.more-trigger:hover),
  :global(.more-trigger[data-state='open']) {
    color: var(--ink);
    background: var(--rail);
  }
</style>
