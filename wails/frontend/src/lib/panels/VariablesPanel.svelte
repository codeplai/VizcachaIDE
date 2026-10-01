<script lang="ts">
  import { t } from '../i18n'
  import { currentFrame, debugState } from '../stores'
</script>

<section class="block">
  <h3 class="side-h">
    {$t('panels.variables', { values: { function: $currentFrame?.function ?? '' } })}
  </h3>
  <div class="vars">
    {#each $debugState?.variables ?? [] as variable (variable.name)}
      <div class="var" class:changed={variable.changed}>
        <span class="n">{variable.name}</span>
        <span class="t">
          {variable.changed
            ? `${variable.typeName} · ${$t('panels.justChanged')}`
            : variable.typeName}
        </span>
        <span class="v">{variable.value}</span>
      </div>
    {:else}
      <p class="empty">{$t('empty.variables')}</p>
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
  .vars {
    display: grid;
    gap: 6px;
  }
  .var {
    background: var(--win);
    border: 1px solid var(--line);
    border-radius: 8px;
    padding: 7px 10px;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 2px 8px;
    font-size: 13px;
  }
  .var.changed {
    border-color: var(--sand);
    box-shadow: inset 3px 0 0 var(--sand);
  }
  .n {
    font: 600 13px var(--mono);
  }
  .t {
    font: 400 11.5px var(--mono);
    color: var(--muted);
    grid-column: 1;
  }
  .v {
    font: 600 15px var(--mono);
    color: var(--go);
    grid-row: 1 / span 2;
    grid-column: 2;
    align-self: center;
  }
  .empty {
    margin: 0;
    color: var(--muted);
    font-size: 13px;
  }
</style>
