<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    isValidProjectName,
    packageBusy,
    runPackageAction,
    suggestedProjectName
  } from '../stores'

  let projectName = $state('')
  let touchedName = false

  // Prefill with the folder's name until the person types their own.
  $effect(() => {
    if (!touchedName) projectName = $suggestedProjectName
  })

  const nameOk = $derived(isValidProjectName(projectName))
</script>

<form
  class="field"
  onsubmit={(event) => {
    event.preventDefault()
    if (nameOk && !$packageBusy) void runPackageAction(bridge, 'init', projectName)
  }}
>
  <label for="module-name">{$t('shell.modulesNewName')}</label>
  <div class="row">
    <input
      id="module-name"
      type="text"
      bind:value={projectName}
      oninput={() => (touchedName = true)}
      aria-invalid={!nameOk}
      aria-describedby="module-name-hint"
    />
    <button class="dlg-button primary" type="submit" disabled={!nameOk || $packageBusy}>
      {$t('shell.modulesCreate')}
    </button>
  </div>
  <span id="module-name-hint" class="hint" class:bad={!nameOk}>
    {nameOk ? $t('shell.modulesNameHint') : $t('shell.modulesNameInvalid')}
  </span>
</form>

<style>
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
</style>
