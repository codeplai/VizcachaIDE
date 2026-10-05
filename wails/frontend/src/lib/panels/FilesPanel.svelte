<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import {
    closeFolder,
    fileTree,
    folderOf,
    openFolder,
    refreshTree,
    selectedNode,
    startCreate
  } from '../stores'
  import FileTreeMenu from './FileTreeMenu.svelte'
  import FileTreeNode from './FileTreeNode.svelte'

  /** Only a click on the empty area opens the root's menu; rows have their own. */
  const onlyBackground =
    (props: Record<string, unknown>) =>
    (event: MouseEvent): void => {
      if (event.target !== event.currentTarget) return
      ;(props.oncontextmenu as (event: MouseEvent) => void)?.(event)
    }
</script>

{#snippet area(props: Record<string, unknown>)}
  <div {...props} class="tree-area" oncontextmenu={onlyBackground(props)}>
    <ul class="tree" aria-label={$t('a11y.fileTree')}>
      <FileTreeNode node={$fileTree!} />
    </ul>
  </div>
{/snippet}

<section class="side files">
  <div class="head">
    <h2 class="side-h">{$t('panels.files')}</h2>
    {#if $fileTree}
      <div class="tools">
        <button
          type="button"
          class="tool"
          title={$t('tree.newFile')}
          aria-label={$t('tree.newFile')}
          onclick={() => startCreate('file', folderOf($selectedNode))}
        >
          <svg viewBox="0 0 16 16" aria-hidden="true"
            ><path d="M4 1.5h5l3.5 3.5v9.5H4z M9 1.5V5h3.5 M8 8v4 M6 10h4" /></svg
          >
        </button>
        <button
          type="button"
          class="tool"
          title={$t('tree.newFolder')}
          aria-label={$t('tree.newFolder')}
          onclick={() => startCreate('folder', folderOf($selectedNode))}
        >
          <svg viewBox="0 0 16 16" aria-hidden="true"
            ><path d="M1.5 3.5h4.5l1.5 1.5h7v8h-13z M8 7v4 M6 9h4" /></svg
          >
        </button>
        <button
          type="button"
          class="tool"
          title={$t('tree.refresh')}
          aria-label={$t('tree.refresh')}
          onclick={() => void refreshTree(bridge)}
        >
          <svg viewBox="0 0 16 16" aria-hidden="true"
            ><path d="M13 8a5 5 0 1 1-1.5-3.5 M13 2.5v3h-3" /></svg
          >
        </button>
        <button
          type="button"
          class="tool"
          title={$t('shell.closeFolder')}
          aria-label={$t('shell.closeFolder')}
          onclick={() => void closeFolder(bridge)}
        >
          <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M4 4l8 8 M12 4l-8 8" /></svg>
        </button>
      </div>
    {/if}
  </div>
  {#if $fileTree}
    <FileTreeMenu path={null} row={area} />
  {:else}
    <p class="empty">{$t('empty.files')}</p>
    <button type="button" class="mini cta" onclick={() => openFolder(bridge)}>
      {$t('empty.filesAction')}
    </button>
  {/if}
</section>

<style>
  .files {
    align-content: stretch;
    grid-template-rows: auto 1fr;
    /* Long file names must not widen the column and push the header buttons out of view. */
    grid-template-columns: minmax(0, 1fr);
    padding-top: 0;
  }
  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 8px 4px 0;
    /* The buttons stay reachable while the tree scrolls. */
    position: sticky;
    top: 0;
    z-index: 1;
    background: var(--chrome);
  }
  .tools {
    display: flex;
    gap: 2px;
  }
  .tool {
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--muted);
    cursor: pointer;
  }
  .tool:hover {
    background: var(--rail);
    color: var(--ink);
  }
  .tool svg {
    width: 15px;
    height: 15px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.3;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .tree-area {
    min-height: 100%;
  }
  .tree {
    margin: 0;
    padding: 0;
  }
  .cta {
    justify-self: start;
    margin: 6px 14px;
  }
</style>
