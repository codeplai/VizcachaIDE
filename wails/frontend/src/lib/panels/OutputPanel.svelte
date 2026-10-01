<script lang="ts">
  import { formatSeconds, locale, t } from '../i18n'
  import { outputLines, type OutputLine } from '../stores'

  const valuesOf = (line: OutputLine, language: string | null | undefined) => {
    const values: Record<string, string | number> = { ...line.values }
    if (line.seconds !== undefined) values.seconds = formatSeconds(language ?? 'en', line.seconds)
    return values
  }
</script>

<div class="term">
  {#each $outputLines as line, index (index)}
    <div class={line.tone}>
      {#if line.key}{$t(line.key, { values: valuesOf(line, $locale) })}{:else}{line.text}{/if}
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
</style>
