<script lang="ts">
  import { bridge } from '../bridge'
  import { resolveLanguage, systemLanguage } from '../language'
  import { t } from '../i18n'
  import {
    cursor,
    debugActive,
    lspStatus,
    settings,
    toolStatus,
    tools,
    updateSettings
  } from '../stores'

  const lspKey = {
    starting: 'status.lspStarting',
    ready: 'status.lspReady',
    unavailable: 'status.lspUnavailable'
  } as const

  const language = $derived(resolveLanguage($settings?.language ?? 'auto', systemLanguage()))
  const otherLanguage = $derived(language === 'es' ? 'en' : 'es')
  const goVersion = $derived(toolStatus('go', $tools)?.version ?? '')
</script>

<footer class="status" class:debug={$debugActive}>
  {#if $debugActive}
    <span>{$t('status.debugging')}</span>
  {:else}
    <span
      class="live"
      data-status={$lspStatus}
      title={$lspStatus === 'unavailable' ? $t('errors.goplsNotFound') : undefined}
    >
      {$t(lspKey[$lspStatus])}
    </span>
  {/if}
  {#if goVersion}<span>{$t('status.goVersion', { values: { version: goVersion } })}</span>{/if}
  <span class="sp">
    {$t('status.position', { values: { line: $cursor.line, column: $cursor.column } })}
  </span>
  <button type="button" onclick={() => updateSettings(bridge, { language: otherLanguage })}>
    {$t(`language.${language}`)}
  </button>
</footer>

<style>
  .status {
    height: 28px;
    flex: none;
    display: flex;
    gap: 18px;
    align-items: center;
    padding: 0 14px;
    font-size: 12px;
    color: var(--muted);
    background: var(--chrome);
    border-top: 1px solid var(--line);
  }
  .status.debug {
    background: var(--sand-soft);
    color: var(--ink);
  }
  .sp {
    margin-left: auto;
  }
  .live {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }
  .live::before {
    content: '';
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--ok);
  }
  .live[data-status='starting']::before {
    background: var(--sand);
  }
  .live[data-status='unavailable']::before {
    background: var(--err);
  }
  button {
    border: 0;
    background: none;
    font-size: 12px;
    color: inherit;
    cursor: pointer;
    padding: 0 2px;
  }
  button:hover {
    color: var(--ink);
    text-decoration: underline;
  }
</style>
