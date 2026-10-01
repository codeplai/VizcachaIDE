<script lang="ts">
  import { bridge } from '../bridge'
  import type { ExplainedDiagnostic } from '../domain'
  import { t } from '../i18n'
  import { goToLocation } from '../stores'

  let { item }: { item: ExplainedDiagnostic } = $props()

  const location = $derived(item.diagnostic.location)
  const searchUrl = $derived(
    `https://www.google.com/search?q=${encodeURIComponent(`golang ${item.diagnostic.message}`)}`
  )
</script>

<article class="card">
  <h3>{item.explanation?.title ?? item.diagnostic.message}</h3>
  {#if item.explanation}
    <p>{item.explanation.body}</p>
    <div class="fix"><b>{$t('assistant.tryThis')}</b> {item.explanation.fixHint}</div>
  {/if}
  <div class="row">
    {#if location}
      <button type="button" class="mini go" onclick={() => goToLocation(bridge, location)}>
        {$t('assistant.goToLine', { values: { line: location.line } })}
      </button>
    {/if}
    <button type="button" class="mini" onclick={() => bridge.system.openUrl(searchUrl)}>
      {$t('assistant.searchError')}
    </button>
  </div>
  <div class="raw">{item.diagnostic.rawText}</div>
</article>

<style>
  .card {
    background: var(--win);
    border: 1px solid var(--line);
    border-radius: 10px;
    padding: 12px;
    display: grid;
    gap: 8px;
  }
  h3 {
    margin: 0;
    font-size: 15px;
    line-height: 1.3;
  }
  p {
    margin: 0;
    font-size: 13.5px;
    color: var(--muted);
  }
  .fix {
    font-size: 13.5px;
    background: var(--go-soft);
    border-radius: 7px;
    padding: 8px 10px;
  }
  .raw {
    font: 400 12px var(--mono);
    color: var(--muted);
    border-top: 1px dashed var(--line);
    padding-top: 8px;
    overflow-wrap: anywhere;
  }
  .row {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
</style>
