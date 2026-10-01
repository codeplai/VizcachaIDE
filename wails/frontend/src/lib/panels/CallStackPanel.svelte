<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { baseName, debugState, goToLocation } from '../stores'
</script>

<ul class="stack" aria-label={$t('panels.callStack')}>
  {#each $debugState?.frames ?? [] as frame, index (frame.frameId)}
    {@const location = frame.location}
    <li>
      <button
        type="button"
        class:on={index === 0}
        disabled={!location}
        onclick={() => location && goToLocation(bridge, location)}
      >
        {frame.function}
        {#if location}· {baseName(location.file)}:{location.line}{/if}
      </button>
    </li>
  {:else}
    <li class="empty">{$t('panels.noFrames')}</li>
  {/each}
</ul>

<style>
  .stack {
    display: grid;
    gap: 4px;
    margin: 0;
    padding: 0;
    font: 400 12.5px var(--mono);
  }
  li {
    list-style: none;
  }
  button {
    width: 100%;
    text-align: left;
    border: 0;
    background: none;
    font: inherit;
    padding: 4px 8px;
    border-radius: 6px;
    color: var(--muted);
    cursor: pointer;
  }
  button:hover:not(:disabled) {
    background: var(--rail);
  }
  button.on {
    background: var(--sand-soft);
    color: var(--ink);
    font-weight: 600;
  }
  .empty {
    color: var(--muted);
  }
</style>
