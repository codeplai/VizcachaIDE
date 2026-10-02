<script lang="ts">
  import { t } from '../i18n'
  import { currentFrame, debugState, isHiddenVariable, shortFunctionName } from '../stores'
  import VariableCard from './VariableCard.svelte'
</script>

<section class="block">
  <h3 class="side-h">
    {$t('panels.variables', {
      values: { function: shortFunctionName($currentFrame?.function ?? '') }
    })}
  </h3>
  <div class="vars">
    {#each ($debugState?.variables ?? []).filter((item) => !isHiddenVariable(item)) as variable, position (position)}
      <VariableCard {variable} />
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
  .vars {
    display: grid;
    gap: 6px;
  }
  .empty {
    margin: 0;
    color: var(--muted);
    font-size: 13px;
  }
</style>
