<script lang="ts">
  import { t } from '../i18n'
  import { fileTree, matchCount, searchState } from '../stores'

  const total = $derived(matchCount($searchState.files))
</script>

<!-- A live region: screen readers hear the count, an error, or what Replace changed. -->
<p class="status" role="status" aria-live="polite" class:error={$searchState.status === 'error'}>
  {#if !$fileTree}
    {$t('search.noFolder')}
  {:else if $searchState.status === 'error'}
    {$t($searchState.error)}
  {:else if $searchState.status === 'done'}
    {#if total === 0}
      {$t('search.none')}
    {:else}
      {$t('search.summary', { values: { count: total, files: $searchState.files.length } })}
      {#if $searchState.truncated}{$t('search.truncated')}{/if}
    {/if}
  {/if}
  {#if $searchState.replaced}
    {$t('search.replaced', {
      values: { count: $searchState.replaced.matches, files: $searchState.replaced.files }
    })}
  {/if}
</p>

<style>
  .status {
    margin: 6px 14px;
    font-size: 12px;
    color: var(--muted);
  }
  .status.error {
    color: var(--danger, #b3261e);
  }
</style>
