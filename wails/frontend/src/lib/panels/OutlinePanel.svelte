<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { activePath, outline, refreshOutline } from '../stores'

  $effect(() => {
    void refreshOutline(bridge, $activePath)
  })
</script>

<section class="side">
  <h2 class="side-h">{$t('panels.outline')}</h2>
  {#each $outline as symbol (symbol.name + symbol.location.line)}
    <div class="symbol">
      <span class="kind">{symbol.kind === 'function' ? 'ƒ' : '•'}</span>
      <span class="name">{symbol.name}</span>
      <span class="line">{symbol.location.line}</span>
    </div>
  {:else}
    <p class="empty">{$t('empty.outline')}</p>
  {/each}
</section>

<style>
  .side {
    height: 100%;
    background: var(--chrome);
    padding: 10px 0;
    overflow: auto;
    display: grid;
    align-content: start;
    gap: 2px;
  }
  .side-h {
    margin: 0;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
    padding: 2px 14px 6px;
  }
  .symbol {
    display: flex;
    gap: 8px;
    align-items: baseline;
    padding: 4px 14px;
    font: 400 13px var(--mono);
  }
  .kind {
    color: var(--code-fn);
  }
  .line {
    margin-left: auto;
    color: var(--muted);
    font-size: 11px;
  }
  .empty {
    margin: 0;
    padding: 4px 14px;
    color: var(--muted);
    font-size: 13px;
  }
</style>
