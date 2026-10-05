<script lang="ts">
  import { onMount } from 'svelte'
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { detectPlatform } from '../platform'
  import { activeTerminal, closeTerminal, startTerminal, terminalSessions } from '../stores'
  import TerminalView from './TerminalView.svelte'

  let { visible }: { visible: boolean } = $props()

  const shellName = $derived(detectPlatform() === 'windows' ? 'PowerShell' : $t('terminal.shell'))

  // The first time the tab is shown it opens a shell in the open folder.
  onMount(() => {
    if ($terminalSessions.length === 0) void startTerminal(bridge)
  })
</script>

<div class="terminal">
  <div class="bar">
    <div class="sessions" role="tablist" aria-label={$t('terminal.sessions')}>
      {#each $terminalSessions as session (session.id)}
        <button
          type="button"
          role="tab"
          class="session"
          class:on={$activeTerminal === session.id}
          aria-selected={$activeTerminal === session.id}
          onclick={() => activeTerminal.set(session.id)}
        >
          {$t('terminal.session', { values: { number: session.number, shell: shellName } })}
        </button>
      {/each}
    </div>
    <button
      type="button"
      class="action new"
      title={$t('terminal.newTip')}
      onclick={() => void startTerminal(bridge)}
    >
      {$t('terminal.new')}
    </button>
    <button
      type="button"
      class="action kill"
      title={$t('terminal.killTip')}
      disabled={!$activeTerminal}
      onclick={() => $activeTerminal && void closeTerminal(bridge, $activeTerminal)}
    >
      {$t('terminal.kill')}
    </button>
  </div>
  <div class="views">
    {#each $terminalSessions as session (session.id)}
      <TerminalView {session} active={visible && $activeTerminal === session.id} />
    {/each}
  </div>
</div>

<style>
  .terminal {
    height: 100%;
    display: grid;
    grid-template-rows: 28px 1fr;
    min-height: 0;
    background: var(--win);
  }
  .bar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 0 8px;
    border-bottom: 1px solid var(--line);
  }
  .sessions {
    display: flex;
    gap: 2px;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
  }
  .session,
  .action {
    border: 0;
    background: none;
    cursor: pointer;
    font-size: 12px;
    color: var(--muted);
    padding: 3px 8px;
    border-radius: 5px;
    white-space: nowrap;
  }
  .session.on {
    background: var(--go-soft);
    color: var(--ink);
    font-weight: 600;
  }
  .session:hover,
  .action:hover:not(:disabled) {
    color: var(--ink);
  }
  .action:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .views {
    position: relative;
    min-height: 0;
  }
</style>
