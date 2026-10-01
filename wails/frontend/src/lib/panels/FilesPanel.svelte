<script lang="ts">
  import { bridge } from '../bridge'
  import type { FileNode } from '../domain'
  import { t } from '../i18n'
  import { activePath, fileTree, openFile, openFolder, problemCount } from '../stores'

  const flatten = (node: FileNode): FileNode[] => node.children
</script>

<section class="side">
  <h2 class="side-h">{$t('panels.files')}</h2>
  {#if $fileTree}
    <div class="file dir">▾ {$fileTree.name}</div>
    {#each flatten($fileTree) as file (file.path)}
      <button
        type="button"
        class="file"
        class:on={file.path === $activePath}
        onclick={() => openFile(bridge, file.path)}
      >
        {file.name}
        {#if file.path === $activePath && $problemCount > 0}<span class="dot"></span>{/if}
      </button>
    {/each}
  {:else}
    <p class="empty">{$t('empty.files')}</p>
    <button type="button" class="cta" onclick={() => openFolder(bridge)}>
      {$t('empty.filesAction')}
    </button>
  {/if}
</section>

<style>
  .side {
    height: 100%;
    background: var(--chrome);
    padding: 10px 0;
    overflow: auto;
    display: grid;
    align-content: start;
    gap: 2px;
  }
  .side-h {
    margin: 0;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--muted);
    padding: 2px 14px 6px;
  }
  .file {
    border: 0;
    background: none;
    text-align: left;
    padding: 4px 14px 4px 22px;
    color: var(--muted);
    font-size: 13px;
    display: flex;
    gap: 8px;
    align-items: center;
    cursor: pointer;
  }
  .file.dir {
    padding-left: 14px;
    color: var(--ink);
    font-weight: 600;
    cursor: default;
  }
  .file.on {
    background: var(--go-soft);
    color: var(--ink);
    font-weight: 600;
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--err);
    margin-left: auto;
  }
  .empty {
    margin: 0;
    padding: 4px 14px;
    color: var(--muted);
    font-size: 13px;
  }
  .cta {
    justify-self: start;
    margin: 6px 14px;
    padding: 5px 10px;
    border: 1px solid var(--line);
    border-radius: 7px;
    background: var(--win);
    font: 600 12.5px var(--ui);
    cursor: pointer;
  }
</style>
