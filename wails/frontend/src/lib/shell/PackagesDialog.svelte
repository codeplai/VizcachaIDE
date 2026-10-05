<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    activeCodeLanguage,
    capabilities,
    hasProject,
    lastRunConfiguration,
    openDialog,
    packageBusy,
    packagesFolder,
    runPackageAction
  } from '../stores'
  import Modal from './Modal.svelte'
  import PackagesAddForm from './PackagesAddForm.svelte'
  import PackagesInitForm from './PackagesInitForm.svelte'
  import PackagesResult from './PackagesResult.svelte'

  const actions = $derived($capabilities?.packageActions ?? [])
  const isGo = $derived($activeCodeLanguage === 'go')
  const isPython = $derived($activeCodeLanguage === 'python')
  const isRust = $derived($activeCodeLanguage === 'rust')
  const hint = $derived.by(() => {
    if (isGo) return $t('shell.modulesIntro')
    if (isPython) return $t('packages.pipHint')
    return isRust ? $t('packages.cargoHint') : undefined
  })
  const project = $derived.by(() => {
    const found = $lastRunConfiguration?.project
    return found && found.kind !== 'folder' ? found : null
  })

  const needsInit = $derived(actions.includes('init') && !$hasProject)
</script>

<Modal
  open={$openDialog === 'packages'}
  title={$t(isGo ? 'shell.goModules' : 'packages.title')}
  description={hint}
  onClose={() => openDialog.set(null)}
>
  {#if project && isGo}
    <dl>
      <dt>{$t('shell.modulesName')}</dt>
      <dd>{project.name}</dd>
      <dt>{$t('shell.modulesFolder')}</dt>
      <dd>{project.root}</dd>
    </dl>
  {:else if isGo && !$hasProject}
    <p class="none">{$t('shell.modulesNone')}</p>
  {/if}

  {#if !$packagesFolder}
    <p>{$t(isGo ? 'shell.modulesNoFolder' : 'packages.noFolder')}</p>
  {:else if needsInit}
    <PackagesInitForm />
  {:else}
    {#if actions.includes('tidy')}
      <div class="field">
        <span class="label">{$t('shell.modulesTidyTitle')}</span>
        <div class="row">
          <button
            class="dlg-button"
            type="button"
            disabled={$packageBusy}
            onclick={() => runPackageAction(bridge, 'tidy')}
          >
            {$t('shell.modulesTidy')}
          </button>
          <span class="hint">{$t('shell.modulesTidyHint')}</span>
        </div>
      </div>
    {/if}
    {#if actions.includes('add')}
      <PackagesAddForm canRemove={actions.includes('remove')} />
    {/if}
    {#if actions.includes('list')}
      <div class="field">
        <button
          class="dlg-button"
          type="button"
          disabled={$packageBusy}
          onclick={() => runPackageAction(bridge, 'list')}
        >
          {$t('packages.list')}
        </button>
      </div>
    {/if}
  {/if}

  <PackagesResult />
</Modal>

<style>
  dl {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 4px 16px;
    margin: 0;
    font-size: 14px;
  }
  dd {
    margin: 0;
    font-family: var(--mono);
    color: var(--muted);
    overflow-wrap: anywhere;
  }
  .row {
    display: flex;
    gap: 8px;
    align-items: center;
  }
</style>
