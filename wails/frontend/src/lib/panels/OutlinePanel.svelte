<script lang="ts">
  import { t } from '../i18n'
  import { outline } from '../stores'
  import OutlineNode from './OutlineNode.svelte'
</script>

<section class="side">
  <h2 class="side-h">{$t('panels.outline')}</h2>
  {#if $outline.length > 0}
    <ul class="symbols">
      {#each $outline as symbol (symbol.name + symbol.location.line)}
        <OutlineNode {symbol} />
      {/each}
    </ul>
  {:else}
    <p class="empty">{$t('empty.outline')}</p>
  {/if}
</section>

<style>
  .symbols {
    margin: 0;
    padding: 0;
  }
</style>
