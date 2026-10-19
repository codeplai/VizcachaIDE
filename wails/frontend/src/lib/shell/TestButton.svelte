<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { activePath, capabilities, running, testActiveFile } from '../stores'
</script>

<!-- Only for the languages that can run the tests of a project (Rust: cargo test). -->
{#if $capabilities?.test}
  <button
    class="btn"
    type="button"
    title={$t('tooltips.test')}
    disabled={!$activePath || $running}
    onclick={() => testActiveFile(bridge)}
  >
    <span>{$t('actions.test')}</span>
  </button>
{/if}

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border-radius: 8px;
    padding: 7px 14px;
    font-weight: 700;
    font-size: 14px;
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    cursor: pointer;
    white-space: nowrap;
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
