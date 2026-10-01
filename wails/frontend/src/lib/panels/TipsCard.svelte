<script lang="ts">
  import { t } from '../i18n'

  const tips = [
    { lead: 'empty.tipRunLead', rest: 'empty.tipRunRest' },
    { lead: 'empty.tipWrongLead', rest: 'empty.tipWrongRest' },
    { lead: 'empty.tipDebugLead', rest: 'empty.tipDebugRest' }
  ]
</script>

<div class="tips">
  {#each tips as tip (tip.lead)}
    <p><b>{$t(tip.lead)}</b> {$t(tip.rest)}</p>
  {/each}
</div>

<style>
  .tips {
    display: grid;
    gap: 10px;
    font-size: 13.5px;
    color: var(--muted);
  }
  p {
    margin: 0;
  }
  b {
    color: var(--ink);
  }
</style>
