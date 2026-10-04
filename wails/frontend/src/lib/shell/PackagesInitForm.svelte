<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    activeCodeLanguage,
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

  const isRust = $derived($activeCodeLanguage === 'rust')
  const nameOk = $derived(isValidProjectName(projectName, $activeCodeLanguage))
  // Cargo names a package, Go a module: each has its own words.
  const texts = $derived(
    isRust
      ? {
          label: 'packages.cargoName',
          create: 'packages.cargoCreate',
          hint: 'packages.cargoHint',
          invalid: 'packages.cargoNameInvalid'
        }
      : {
          label: 'shell.modulesNewName',
          create: 'shell.modulesCreate',
          hint: 'shell.modulesNameHint',
          invalid: 'shell.modulesNameInvalid'
        }
  )
</script>

<form
  class="field"
  onsubmit={(event) => {
    event.preventDefault()
    if (nameOk && !$packageBusy) void runPackageAction(bridge, 'init', projectName)
  }}
>
  <label for="module-name">{$t(texts.label)}</label>
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
      {$t(texts.create)}
    </button>
  </div>
  <span id="module-name-hint" class="hint" class:bad={!nameOk}>
    {$t(nameOk ? texts.hint : texts.invalid)}
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
