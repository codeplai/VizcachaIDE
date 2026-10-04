<script lang="ts">
  import { t } from '../i18n'
  import { openDialog, settingsTab, type SettingsTab } from '../stores'
  import Modal from './Modal.svelte'
  import SettingsEditor from './SettingsEditor.svelte'
  import SettingsGeneral from './SettingsGeneral.svelte'
  import SettingsTools from './SettingsTools.svelte'
  import SettingsUpdates from './SettingsUpdates.svelte'

  const tabs: { id: SettingsTab; label: string }[] = [
    { id: 'general', label: 'settings.general' },
    { id: 'editor', label: 'settings.editor' },
    { id: 'tools', label: 'settings.tools' },
    { id: 'updates', label: 'updates.title' }
  ]
</script>

<Modal
  open={$openDialog === 'settings'}
  title={$t('panels.settings')}
  onClose={() => openDialog.set(null)}
>
  <div class="tabs" role="tablist" aria-label={$t('a11y.settingsTabs')}>
    {#each tabs as tab (tab.id)}
      <button
        type="button"
        role="tab"
        class="tab"
        class:on={$settingsTab === tab.id}
        aria-selected={$settingsTab === tab.id}
        onclick={() => settingsTab.set(tab.id)}
      >
        {$t(tab.label)}
      </button>
    {/each}
  </div>
  <div class="section" role="tabpanel">
    {#if $settingsTab === 'general'}
      <SettingsGeneral />
    {:else if $settingsTab === 'editor'}
      <SettingsEditor />
    {:else if $settingsTab === 'tools'}
      <SettingsTools />
    {:else}
      <SettingsUpdates />
    {/if}
  </div>
</Modal>

<style>
  .tabs {
    display: flex;
    gap: 4px;
    border-bottom: 1px solid var(--line);
  }
  .tab {
    border: 0;
    background: none;
    cursor: pointer;
    padding: 6px 12px;
    font: 600 13.5px var(--ui);
    color: var(--muted);
    border-radius: 7px 7px 0 0;
  }
  .tab.on {
    color: var(--ink);
    box-shadow: inset 0 -2px 0 var(--go);
  }
  .section {
    display: grid;
    gap: 14px;
    min-height: 190px;
    align-content: start;
  }
</style>
