<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    activeFileName,
    activePath,
    debugActive,
    fileTree,
    parentName,
    programArguments,
    runActiveFile,
    running,
    startDebugging,
    stopDebugging,
    stopProgram
  } from '../stores'
  import mascot from '../../assets/brand/mark.png'
  import FileButtons from './FileButtons.svelte'
  import FileMenu from './FileMenu.svelte'
  import MoreMenu from './MoreMenu.svelte'
</script>

<header class="titlebar">
  <div class="brand">
    <img class="brand-mark" src={mascot} alt="" width="28" height="28" />{$t('app.name')}
  </div>
  <FileMenu />
  <div class="crumbs">
    {#if $activePath}
      {$fileTree?.name ?? parentName($activePath)} / <b>{$activeFileName}</b>
    {/if}
  </div>
  <div class="actions">
    <FileButtons />
    <input
      class="args"
      type="text"
      spellcheck="false"
      autocomplete="off"
      placeholder={$t('settings.programArgs')}
      aria-label={$t('settings.programArgs')}
      title={$t('settings.programArgsHint')}
      disabled={$debugActive || $running}
      bind:value={$programArguments}
      onkeydown={(event) => event.key === 'Enter' && $activePath && runActiveFile(bridge)}
    />
    {#if $debugActive}
      <button
        class="btn stop"
        type="button"
        title={$t('tooltips.stopDebugging')}
        onclick={() => stopDebugging(bridge)}
      >
        <span>{$t('actions.stopDebugging')}</span><span class="k">Shift+F5</span>
      </button>
    {:else}
      <button
        class="btn primary"
        type="button"
        title={$t('tooltips.run')}
        disabled={!$activePath}
        onclick={() => runActiveFile(bridge)}
      >
        <span class="tri"></span><span>{$t('actions.run')}</span><span class="k">F5</span>
      </button>
      <button
        class="btn"
        type="button"
        title={$t('tooltips.debug')}
        disabled={!$activePath}
        onclick={() => startDebugging(bridge)}
      >
        <span class="bug"></span><span>{$t('actions.debug')}</span><span class="k">F6</span>
      </button>
      {#if $running}
        <button
          class="btn"
          type="button"
          title={$t('tooltips.stop')}
          onclick={() => stopProgram(bridge)}
        >
          <span>{$t('actions.stop')}</span>
        </button>
      {/if}
    {/if}
    <MoreMenu />
  </div>
</header>

<style>
  .titlebar {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 52px;
    flex: none;
    padding: 0 14px;
    background: var(--chrome);
    border-bottom: 1px solid var(--line);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 700;
    white-space: nowrap;
  }
  .brand-mark {
    width: 28px;
    height: 28px;
    display: block;
    object-fit: contain;
  }
  .crumbs {
    color: var(--muted);
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .crumbs b {
    color: var(--ink);
    font-weight: 600;
  }
  .actions {
    margin-left: auto;
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .args {
    width: clamp(80px, 11vw, 200px);
    min-width: 0;
    padding: 7px 10px;
    border-radius: 8px;
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    font: 500 13px var(--mono);
  }
  .args:disabled {
    opacity: 0.55;
  }
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
  .btn .k {
    font: 600 11px var(--mono);
    color: var(--muted);
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn.primary {
    background: var(--go);
    border-color: var(--go);
    color: var(--win);
  }
  .btn.primary .k {
    color: var(--win);
    opacity: 0.8;
  }
  .btn.stop {
    color: var(--err);
  }
  .tri {
    width: 0;
    height: 0;
    border-left: 9px solid currentColor;
    border-top: 6px solid transparent;
    border-bottom: 6px solid transparent;
  }
  .bug {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 2px solid currentColor;
  }
</style>
