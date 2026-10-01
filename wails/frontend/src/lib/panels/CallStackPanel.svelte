<script lang="ts">
  import { t } from '../i18n'
  import { baseName, debugState } from '../stores'
</script>

<section class="block">
  <h3 class="side-h">{$t('panels.callStack')}</h3>
  <div class="stack">
    {#each $debugState?.frames ?? [] as frame, index (frame.frameId)}
      <div class:on={index === 0}>
        {frame.function}
        {#if frame.location}· {baseName(frame.location.file)}:{frame.location.line}{/if}
      </div>
    {/each}
  </div>
</section>

<style>
  .block {
    display: grid;
    gap: 12px;
  }
  .side-h {
    margin: 0;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
  }
  .stack {
    display: grid;
    gap: 4px;
    font: 400 12.5px var(--mono);
  }
  .stack div {
    padding: 4px 8px;
    border-radius: 6px;
    color: var(--muted);
  }
  .stack div.on {
    background: var(--sand-soft);
    color: var(--ink);
    font-weight: 600;
  }
</style>
