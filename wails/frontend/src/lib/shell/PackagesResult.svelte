<script lang="ts">
  import { t } from '../i18n'
  import { packageBusy, packageOutput, packageTask, runResult } from '../stores'

  // The texts of 2.0 name Go and its verbs ("get" is the add action); other languages share two.
  const specificKey = (kind: 'Done' | 'Failed'): string | null => {
    const task = $packageTask
    if (!task || task.codeLanguage !== 'go') return null
    if (task.action === 'remove' || task.action === 'list') return null
    return `shell.modules${kind}.${task.action === 'add' ? 'get' : task.action}`
  }

  // Go keeps the texts of 2.0 (they name Go); other languages share the neutral ones.
  const textKey = (goKey: string, otherKey: string): string =>
    $packageTask && $packageTask.codeLanguage !== 'go' ? otherKey : goKey

  const finished = $derived($packageTask && !$packageBusy && $runResult ? $runResult : null)
  const failedKey = $derived(
    $packageTask?.error
      ? textKey('shell.modulesStartFailed', 'packages.startFailed')
      : finished && finished.exitCode !== 0
        ? (specificKey('Failed') ?? 'packages.failed')
        : null
  )
  const doneKey = $derived(
    finished?.exitCode === 0 && $packageTask ? (specificKey('Done') ?? 'packages.done') : null
  )
</script>

{#if $packageBusy}
  <p role="status">{$t(textKey('shell.modulesWorking', 'packages.working'))}</p>
{/if}
{#if $packageTask && $packageOutput.length > 0}
  <pre
    class="out"
    role="log"
    aria-label={$t(textKey('shell.modulesOutput', 'packages.output'))}>{$packageOutput.join(
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
