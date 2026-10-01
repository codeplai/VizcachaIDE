<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { resumeDebugging, runToCursor, stepInto, stepOut, stepOver } from '../stores'

  const buttons = [
    { label: 'actions.nextLine', tip: 'tooltips.nextLine', key: 'F7', run: stepOver, main: true },
    { label: 'actions.goIntoFunction', tip: 'tooltips.goIntoFunction', key: 'F8', run: stepInto },
    { label: 'actions.leaveFunction', tip: 'tooltips.leaveFunction', key: 'F9', run: stepOut },
    { label: 'actions.continue', tip: 'tooltips.continue', key: 'Shift+F6', run: resumeDebugging },
    { label: 'actions.runToHere', tip: 'tooltips.runToHere', key: 'Ctrl+F10', run: runToCursor }
  ]
</script>

<div class="debugbar" role="toolbar" aria-label={$t('a11y.debugControls')}>
  {#each buttons as button (button.label)}
    <button
      type="button"
      class="db"
      class:main={button.main}
      title={$t(button.tip)}
      onclick={() => button.run(bridge)}
    >
      <span>{$t(button.label)}</span><small>{button.key}</small>
    </button>
  {/each}
</div>

<style>
  .debugbar {
    position: absolute;
    top: 46px;
    right: 16px;
    z-index: 5;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    max-width: calc(100% - 32px);
    gap: 4px;
    align-items: center;
    background: var(--win);
    border: 1px solid var(--line);
    border-radius: 10px;
    padding: 4px;
    box-shadow: 0 10px 24px -14px rgba(8, 18, 28, 0.55);
  }
  .db {
    border: 0;
    background: none;
    cursor: pointer;
    font: 600 12.5px var(--ui);
    padding: 6px 10px;
    border-radius: 7px;
    color: var(--ink);
    display: grid;
    line-height: 1.2;
    text-align: left;
    white-space: nowrap;
  }
  .db small {
    font-weight: 400;
    color: var(--muted);
    font-size: 11px;
  }
  .db.main {
    background: var(--go);
    color: var(--win);
  }
  .db.main small {
    color: var(--win);
    opacity: 0.8;
  }
</style>
