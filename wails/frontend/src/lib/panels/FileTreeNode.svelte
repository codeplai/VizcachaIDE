<script lang="ts">
  import { bridge } from '../bridge'
  import type { FileNode } from '../domain'
  import { t } from '../i18n'
  import {
    activePath,
    collapsedFolders,
    filesWithProblems,
    openFile,
    selectedNode,
    toggleFolder
  } from '../stores'
  import FileTreeNode from './FileTreeNode.svelte'

  let { node, depth = 0 }: { node: FileNode; depth?: number } = $props()

  const open = $derived(!$collapsedFolders.has(node.path))
  const indent = $derived(`${14 + depth * 12}px`)
  const hasProblems = $derived(!node.isDir && $filesWithProblems.includes(node.path))

  const onKeydown = (event: KeyboardEvent): void => {
    if (event.key !== 'Enter' || node.isDir) return
    event.preventDefault()
    void openFile(bridge, node.path)
  }
</script>

<li>
  {#if node.isDir}
    <button
      type="button"
      class="file dir"
      style:padding-left={indent}
      aria-expanded={open}
      aria-label={$t(open ? 'panels.collapseFolder' : 'panels.expandFolder', {
        values: { name: node.name }
      })}
      onclick={() => toggleFolder(node.path)}
    >
      <span aria-hidden="true">{open ? '▾' : '▸'}</span>
      {node.name}
    </button>
    {#if open && node.children.length > 0}
      <ul>
        {#each node.children as child (child.path)}
          <FileTreeNode node={child} depth={depth + 1} />
        {/each}
      </ul>
    {/if}
  {:else}
    <button
      type="button"
      class="file"
      class:on={node.path === $activePath}
      class:sel={node.path === $selectedNode}
      style:padding-left={indent}
      aria-current={node.path === $activePath ? 'true' : undefined}
      onclick={() => selectedNode.set(node.path)}
      ondblclick={() => openFile(bridge, node.path)}
      onkeydown={onKeydown}
    >
      {node.name}
      {#if hasProblems}
        <span class="dot" role="img" aria-label={$t('panels.hasProblems')}></span>
      {/if}
    </button>
  {/if}
</li>

<style>
  li {
    list-style: none;
  }
  ul {
    margin: 0;
    padding: 0;
  }
  .file {
    width: 100%;
    border: 0;
    background: none;
    text-align: left;
    padding: 4px 14px;
    color: var(--muted);
    font-size: 13px;
    display: flex;
    gap: 8px;
    align-items: center;
    cursor: pointer;
  }
  .file:hover {
    background: var(--rail);
  }
  .file.dir {
    color: var(--ink);
    font-weight: 600;
  }
  .file.sel {
    background: var(--rail);
    color: var(--ink);
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
</style>
