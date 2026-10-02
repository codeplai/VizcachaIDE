<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    hasGoMod,
    isValidModuleName,
    isValidPackage,
    lastRunConfiguration,
    moduleBusy,
    moduleOutput,
    modulesFolder,
    moduleTask,
    openDialog,
    runModuleAction,
    runResult,
    suggestedModuleName
  } from '../stores'
  import Modal from './Modal.svelte'

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
  const finished = $derived($moduleTask && !$moduleBusy && $runResult ? $runResult : null)
  const failedKey = $derived(
    $moduleTask?.error
      ? 'shell.modulesStartFailed'
      : finished && finished.exitCode !== 0
        ? `shell.modulesFailed.${$moduleTask?.action}`
        : null
  )
  const doneKey = $derived(
    finished?.exitCode === 0 && $moduleTask ? `shell.modulesDone.${$moduleTask.action}` : null
  )
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

  {#if $moduleBusy}
    <p role="status">{$t('shell.modulesWorking')}</p>
  {/if}
  {#if $moduleTask && $moduleOutput.length > 0}
    <pre class="out" role="log" aria-label={$t('shell.modulesOutput')}>{$moduleOutput.join(
        '\n'
      )}</pre>
  {/if}
  {#if failedKey}
    <p class="bad" role="alert">{$t(failedKey)}</p>
  {:else if doneKey}
    <p class="ok" role="status">{$t(doneKey)}</p>
  {/if}
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
  .ok {
    color: var(--ok);
  }
  .out {
    margin: 0;
    max-height: 180px;
    overflow: auto;
    padding: 8px 10px;
    border-radius: 7px;
    background: var(--chrome);
    border: 1px solid var(--line);
    font: 400 12.5px/1.5 var(--mono);
    white-space: pre-wrap;
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
