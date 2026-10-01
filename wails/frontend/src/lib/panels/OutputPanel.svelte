<script lang="ts">
  import { tick } from 'svelte'
  import { bridge } from '../bridge'
  import { formatSeconds, locale, t } from '../i18n'
  import { goToLocation, outputLines, type OutputLine } from '../stores'

  let terminal: HTMLDivElement | undefined = $state()

  const valuesOf = (line: OutputLine, language: string | null | undefined) => {
    const values: Record<string, string | number> = { ...line.values }
    if (line.seconds !== undefined) values.seconds = formatSeconds(language ?? 'en', line.seconds)
    return values
  }

  // Follow the program: keep the newest line in view.
  $effect(() => {
    void $outputLines
    void tick().then(() => {
      if (terminal) terminal.scrollTop = terminal.scrollHeight
    })
  })
</script>

<div class="term" role="log" aria-live="polite" aria-label={$t('a11y.output')} bind:this={terminal}>
  {#each $outputLines as line, index (index)}
    <div class={line.tone}>
      {#if line.key}
        {$t(line.key, { values: valuesOf(line, $locale) })}
      {:else}
        {#each line.segments ?? [] as segment, position (position)}
          {#if segment.location}
            {@const location = segment.location}
            <button
              type="button"
              class="place"
              title={$t('a11y.goToPlace', { values: { place: segment.text } })}
              onclick={() => goToLocation(bridge, location)}
            >
              {segment.text}
            </button>
          {:else}
            {segment.text}
          {/if}
        {/each}
      {/if}
    </div>
  {:else}
    <div class="system">{$t('empty.output')}</div>
  {/each}
</div>

<style>
  .term {
    font: 400 13px/1.6 var(--mono);
    padding: 8px 14px;
    overflow: auto;
    height: 100%;
    white-space: pre-wrap;
  }
  .system {
    color: var(--muted);
  }
  .error {
    color: var(--err);
  }
  .success {
    color: var(--ok);
  }
  .place {
    border: 0;
    background: none;
    padding: 0;
    font: inherit;
    color: var(--go);
    text-decoration: underline;
    cursor: pointer;
  }
  .place:hover {
    background: var(--go-soft);
  }
</style>
