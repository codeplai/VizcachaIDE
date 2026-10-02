<script lang="ts">
  import { t } from '../i18n'
  import { moduleBusy, moduleOutput, moduleTask, runResult } from '../stores'

  const finished = $derived($moduleTask && !$moduleBusy && $runResult ? $runResult : null)
  const failedKey = $derived(
    $moduleTask?.error
      ? 'shell.modulesStartFailed'
      : finished && finished.exitCode !== 0
        ? `shell.modulesFailed.${$moduleTask?.action}`
        : null
  )
  const doneKey = $derived(
    finished?.exitCode === 0 && $moduleTask ? `shell.modulesDone.${$moduleTask.action}` : null
  )
</script>

{#if $moduleBusy}
  <p role="status">{$t('shell.modulesWorking')}</p>
{/if}
{#if $moduleTask && $moduleOutput.length > 0}
  <pre class="out" role="log" aria-label={$t('shell.modulesOutput')}>{$moduleOutput.join(
      '\n'
    )}</pre>
{/if}
{#if failedKey}
  <p class="bad" role="alert">{$t(failedKey)}</p>
{:else if doneKey}
  <p class="ok" role="status">{$t(doneKey)}</p>
{/if}

<style>
  .bad {
    color: var(--err);
  }
  .ok {
    color: var(--ok);
  }
  .out {
    margin: 0;
    max-height: 180px;
    overflow: auto;
    padding: 8px 10px;
    border-radius: 7px;
    background: var(--chrome);
    border: 1px solid var(--line);
    font: 400 12.5px/1.5 var(--mono);
    white-space: pre-wrap;
  }
</style>
