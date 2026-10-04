<script lang="ts">
  import { t } from '../i18n'
  import {
    assistantOpen,
    assistantProblemCount,
    sidebarOpen,
    toggleAssistant,
    toggleSidebar
  } from '../stores'
</script>

<div class="toggles" role="group" aria-label={$t('a11y.panels')}>
  <button
    type="button"
    class="toggle"
    class:on={$sidebarOpen}
    aria-pressed={$sidebarOpen}
    aria-label={$t('actions.toggleSidebar')}
    title={$t('actions.toggleSidebar')}
    onclick={toggleSidebar}
  >
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <rect x="1.5" y="2.5" width="13" height="11" rx="1.5" />
      <rect class="fill" x="1.5" y="2.5" width="4.5" height="11" rx="1.5" />
    </svg>
  </button>
  <button
    type="button"
    class="toggle"
    class:on={$assistantOpen}
    aria-pressed={$assistantOpen}
    aria-label={$t('actions.toggleAssistant')}
    title={$t('actions.toggleAssistant')}
    onclick={toggleAssistant}
  >
    <svg viewBox="0 0 16 16" aria-hidden="true">
      <rect x="1.5" y="2.5" width="13" height="11" rx="1.5" />
      <rect class="fill" x="10" y="2.5" width="4.5" height="11" rx="1.5" />
    </svg>
    {#if !$assistantOpen && $assistantProblemCount > 0}
      <span class="badge">{$assistantProblemCount}</span>
    {/if}
  </button>
</div>

<style>
  .toggles {
    display: flex;
    gap: 2px;
  }
  .toggle {
    position: relative;
    width: 30px;
    height: 30px;
    border: 0;
    border-radius: 8px;
    background: none;
    color: var(--muted);
    display: grid;
    place-items: center;
    cursor: pointer;
  }
  .toggle:hover {
    background: var(--win);
  }
  svg {
    width: 16px;
    height: 16px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.2;
  }
  svg .fill {
    fill: none;
  }
  .toggle.on {
    color: var(--go);
  }
  .toggle.on svg .fill {
    fill: currentColor;
    opacity: 0.35;
  }
  .badge {
    position: absolute;
    top: 1px;
    right: 1px;
    min-width: 14px;
    height: 14px;
    padding: 0 3px;
    border-radius: 7px;
    background: var(--err);
    color: #fff;
    font: 600 9.5px/14px var(--ui, sans-serif);
  }
</style>
