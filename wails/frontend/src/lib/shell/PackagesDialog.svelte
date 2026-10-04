<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    activeCodeLanguage,
    capabilities,
    hasProject,
    isValidPackage,
    lastRunConfiguration,
    openDialog,
    packageBusy,
    packagesFolder,
    runPackageAction
  } from '../stores'
  import Modal from './Modal.svelte'
  import PackagesInitForm from './PackagesInitForm.svelte'
  import PackagesResult from './PackagesResult.svelte'

  const PKG_GO_DEV = 'https://pkg.go.dev'

  let packageName = $state('')

  const actions = $derived($capabilities?.packageActions ?? [])
  const isGo = $derived($activeCodeLanguage === 'go')
  const isPython = $derived($activeCodeLanguage === 'python')
  const isRust = $derived($activeCodeLanguage === 'rust')
  const hint = $derived.by(() => {
    if (isGo) return $t('shell.modulesIntro')
    if (isPython) return $t('packages.pipHint')
    return isRust ? $t('packages.cargoHint') : undefined
  })
  const placeholder = $derived(isGo ? 'github.com/user/pkg' : isRust ? 'serde' : 'requests')
  const project = $derived.by(() => {
    const found = $lastRunConfiguration?.project
    return found && found.kind !== 'folder' ? found : null
  })

  const invalidKey = $derived(
    isGo ? 'shell.modulesPackageInvalid' : isRust ? 'packages.cargoNameInvalid' : 'packages.invalid'
  )
  const packageOk = $derived(isValidPackage(packageName, $activeCodeLanguage))
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
      <form
        class="field"
        onsubmit={(event) => {
          event.preventDefault()
          if (packageOk && !$packageBusy) void runPackageAction(bridge, 'add', packageName)
        }}
      >
        <label for="module-package">{$t('shell.modulesPackage')}</label>
        <div class="row">
          <input
            id="module-package"
            type="text"
            bind:value={packageName}
            {placeholder}
            aria-invalid={packageName !== '' && !packageOk}
          />
          <button
            class="dlg-button"
            type="submit"
            disabled={packageName.trim() === '' || !packageOk || $packageBusy}
          >
            {$t(isGo ? 'shell.modulesAdd' : 'packages.add')}
          </button>
          {#if actions.includes('remove')}
            <button
              class="dlg-button"
              type="button"
              disabled={packageName.trim() === '' || !packageOk || $packageBusy}
              onclick={() => runPackageAction(bridge, 'remove', packageName)}
            >
              {$t('packages.remove')}
            </button>
          {/if}
        </div>
        {#if packageName !== '' && !packageOk}
          <span class="hint bad">{$t(invalidKey)}</span>
        {/if}
        {#if isGo}
          <button class="link" type="button" onclick={() => bridge.system.openUrl(PKG_GO_DEV)}>
            {$t('shell.modulesSearch')}
          </button>
        {/if}
      </form>
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
  .row input {
    flex: 1;
  }
  .bad {
    color: var(--err);
  }
  .link {
    justify-self: start;
    border: 0;
    background: none;
    padding: 0;
    font: inherit;
    font-size: 13px;
    color: var(--go);
    text-decoration: underline;
    cursor: pointer;
  }
</style>
