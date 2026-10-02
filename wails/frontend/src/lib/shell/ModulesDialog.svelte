<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    hasGoMod,
    isValidModuleName,
    isValidPackage,
    lastRunConfiguration,
    moduleBusy,
    modulesFolder,
    openDialog,
    runModuleAction,
    suggestedModuleName
  } from '../stores'
  import Modal from './Modal.svelte'
  import ModulesResult from './ModulesResult.svelte'

  const PKG_GO_DEV = 'https://pkg.go.dev'

  let moduleName = $state('')
  let packageName = $state('')
  let touchedName = false

  const module = $derived($lastRunConfiguration?.module ?? null)
  // Prefill with the folder's name until the person types their own.
  $effect(() => {
    if (!touchedName) moduleName = $suggestedModuleName
  })

  const nameOk = $derived(isValidModuleName(moduleName))
  const packageOk = $derived(isValidPackage(packageName))
</script>

<Modal
  open={$openDialog === 'modules'}
  title={$t('shell.goModules')}
  description={$t('shell.modulesIntro')}
  onClose={() => openDialog.set(null)}
>
  {#if module}
    <dl>
      <dt>{$t('shell.modulesName')}</dt>
      <dd>{module.modulePath}</dd>
      <dt>{$t('shell.modulesFolder')}</dt>
      <dd>{module.root}</dd>
    </dl>
  {:else if !$hasGoMod}
    <p class="none">{$t('shell.modulesNone')}</p>
  {/if}

  {#if !$modulesFolder}
    <p>{$t('shell.modulesNoFolder')}</p>
  {:else if !$hasGoMod}
    <form
      class="field"
      onsubmit={(event) => {
        event.preventDefault()
        if (nameOk && !$moduleBusy) void runModuleAction(bridge, 'init', moduleName)
      }}
    >
      <label for="module-name">{$t('shell.modulesNewName')}</label>
      <div class="row">
        <input
          id="module-name"
          type="text"
          bind:value={moduleName}
          oninput={() => (touchedName = true)}
          aria-invalid={!nameOk}
          aria-describedby="module-name-hint"
        />
        <button class="dlg-button primary" type="submit" disabled={!nameOk || $moduleBusy}>
          {$t('shell.modulesCreate')}
        </button>
      </div>
      <span id="module-name-hint" class="hint" class:bad={!nameOk}>
        {nameOk ? $t('shell.modulesNameHint') : $t('shell.modulesNameInvalid')}
      </span>
    </form>
  {:else}
    <div class="field">
      <span class="label">{$t('shell.modulesTidyTitle')}</span>
      <div class="row">
        <button
          class="dlg-button"
          type="button"
          disabled={$moduleBusy}
          onclick={() => runModuleAction(bridge, 'tidy')}
        >
          {$t('shell.modulesTidy')}
        </button>
        <span class="hint">{$t('shell.modulesTidyHint')}</span>
      </div>
    </div>
    <form
      class="field"
      onsubmit={(event) => {
        event.preventDefault()
        if (packageOk && !$moduleBusy) void runModuleAction(bridge, 'get', packageName)
      }}
    >
      <label for="module-package">{$t('shell.modulesPackage')}</label>
      <div class="row">
        <input
          id="module-package"
          type="text"
          bind:value={packageName}
          placeholder="github.com/user/pkg"
          aria-invalid={packageName !== '' && !packageOk}
        />
        <button
          class="dlg-button"
          type="submit"
          disabled={packageName.trim() === '' || !packageOk || $moduleBusy}
        >
          {$t('shell.modulesAdd')}
        </button>
      </div>
      {#if packageName !== '' && !packageOk}
        <span class="hint bad">{$t('shell.modulesPackageInvalid')}</span>
      {/if}
      <button class="link" type="button" onclick={() => bridge.system.openUrl(PKG_GO_DEV)}>
        {$t('shell.modulesSearch')}
      </button>
    </form>
  {/if}

  <ModulesResult />
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
