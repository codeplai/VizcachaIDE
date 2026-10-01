<script lang="ts">
  import type { ExplainedDiagnostic } from '../domain'
  import { t } from '../i18n'

  let { item }: { item: ExplainedDiagnostic } = $props()

  const line = $derived(item.diagnostic.location?.line ?? 0)
  const searchUrl = $derived(
    `https://www.google.com/search?q=${encodeURIComponent(`golang ${item.diagnostic.message}`)}`
  )
</script>

<div class="card">
  <h3>{item.explanation?.title ?? item.diagnostic.message}</h3>
  {#if item.explanation}
    <p>{item.explanation.body}</p>
    <div class="fix"><b>{$t('assistant.tryThis')}</b> {item.explanation.fixHint}</div>
  {/if}
  <div class="row">
    <button type="button" class="mini go">
      {$t('assistant.goToLine', { values: { line } })}
    </button>
    <a class="mini" href={searchUrl} target="_blank" rel="noreferrer">
      {$t('assistant.searchError')}
    </a>
  </div>
  <div class="raw">{item.diagnostic.rawText}</div>
</div>

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
  .mini {
    font: 600 12.5px var(--ui);
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    border-radius: 7px;
    padding: 5px 10px;
    text-decoration: none;
    cursor: pointer;
  }
  .mini.go {
    background: var(--go);
    color: var(--win);
    border-color: var(--go);
  }
</style>
