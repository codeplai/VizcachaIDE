<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { baseName, capabilities, debugState, goToLocation } from '../stores'

  const label = $derived($capabilities?.threadsLabel ?? 'debug.goroutines')
  // Go keeps its own words (goroutines); every other language talks about threads.
  const kind = $derived(label === 'debug.goroutines' ? 'goroutine' : 'thread')
</script>

<ul class="routines" aria-label={$t(label)}>
  {#each $debugState?.threads ?? [] as routine (routine.threadId)}
    {@const location = routine.location}
    <li>
      <button
        type="button"
        class:on={routine.threadId === $debugState?.currentThread}
        disabled={!location}
        onclick={() => location && goToLocation(bridge, location)}
      >
        {$t(`panels.${kind}`, { values: { id: routine.threadId } })} · {routine.name}
        {#if location}· {baseName(location.file)}:{location.line}{/if}
      </button>
    </li>
  {:else}
    <li class="empty">{$t(`panels.${kind}sEmpty`)}</li>
  {/each}
</ul>

<style>
  .routines {
    display: grid;
    gap: 4px;
    margin: 0;
    padding: 0;
    font: 400 12.5px var(--mono);
  }
  li {
    list-style: none;
  }
  button {
    width: 100%;
    text-align: left;
    border: 0;
    background: none;
    font: inherit;
    padding: 4px 8px;
    border-radius: 6px;
    color: var(--muted);
    cursor: pointer;
  }
  button:hover:not(:disabled) {
    background: var(--rail);
  }
  button.on {
    background: var(--sand-soft);
    color: var(--ink);
    font-weight: 600;
  }
  button:disabled {
    cursor: default;
  }
  .empty {
    color: var(--muted);
  }
</style>
