<script lang="ts">
  import { onDestroy } from 'svelte'
  import { get } from 'svelte/store'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import type { PackageInfo } from '../domain'
  import {
    activeCodeLanguage,
    createPackageSearch,
    isValidPackage,
    packageBusy,
    runPackageAction
  } from '../stores'
  import PackageSearchList from './PackageSearchList.svelte'

  const PKG_GO_DEV = 'https://pkg.go.dev'
  const LIST_ID = 'package-results'
  const optionId = (index: number): string => `package-option-${index}`

  /** The Add form of the Packages dialog: a field with a search list, Install and Uninstall. */
  let { canRemove }: { canRemove: boolean } = $props()

  let packageName = $state('')
  const search = createPackageSearch(bridge, () => get(activeCodeLanguage))
  onDestroy(search.reset)
  // A search belongs to one language: switching files to another language starts over.
  $effect(() => {
    void $activeCodeLanguage
    search.reset()
  })

  const isGo = $derived($activeCodeLanguage === 'go')
  const isRust = $derived($activeCodeLanguage === 'rust')
  const isCpp = $derived($activeCodeLanguage === 'cpp')
  const placeholder = $derived(
    isGo ? 'github.com/user/pkg' : isRust ? 'serde' : isCpp ? 'fmt' : 'requests'
  )
  const invalidKey = $derived(
    isGo
      ? 'shell.modulesPackageInvalid'
      : isRust
        ? 'packages.cargoNameInvalid'
        : isCpp
          ? 'packages.vcpkgNameInvalid'
          : 'packages.invalid'
  )
  const packageOk = $derived(isValidPackage(packageName, $activeCodeLanguage))
  const listOpen = $derived($search.status === 'done' && $search.results.length > 0)
  const chosenLabel = $derived(
    $search.chosen ? `${$search.chosen.name} ${$search.chosen.version}`.trim() : ''
  )

  const choose = (found: PackageInfo): void => {
    packageName = found.name
    search.choose(found)
  }

  const onKeydown = (event: KeyboardEvent): void => {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      if (!listOpen) return
      event.preventDefault()
      search.move(event.key === 'ArrowDown' ? 1 : -1)
      return
    }
    const option = event.key === 'Enter' && listOpen ? search.activeOption() : null
    if (!option) return
    event.preventDefault()
    choose(option)
  }
</script>

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
      role="combobox"
      autocomplete="off"
      aria-autocomplete="list"
      aria-expanded={listOpen}
      aria-controls={LIST_ID}
      aria-activedescendant={listOpen && $search.active >= 0 ? optionId($search.active) : undefined}
      bind:value={packageName}
      oninput={() => search.typed(packageName)}
      onkeydown={onKeydown}
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
    {#if canRemove}
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
  <PackageSearchList state={$search} listId={LIST_ID} {optionId} onChoose={choose} />
  {#if chosenLabel}
    <span class="hint chosen" role="status"
      >{$t('packages.chosen', { values: { name: chosenLabel } })}</span
    >
  {/if}
  {#if packageName !== '' && !packageOk}
    <span class="hint bad">{$t(invalidKey)}</span>
  {/if}
  {#if isGo}
    <button class="link" type="button" onclick={() => bridge.system.openUrl(PKG_GO_DEV)}>
      {$t('shell.modulesSearch')}
    </button>
  {/if}
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
  .chosen {
    color: var(--go);
    font-weight: 700;
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
