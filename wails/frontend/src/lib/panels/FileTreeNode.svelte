<script lang="ts">
  import { bridge } from '../bridge'
  import type { FileNode } from '../domain'
  import { t } from '../i18n'
  import {
    activePath,
    cancelEdit,
    clearEditError,
    collapsedFolders,
    commitEdit,
    deleteEntry,
    filesWithProblems,
    openFile,
    selectedNode,
    startRename,
    toggleFolder,
    treeEdit
  } from '../stores'
  import FileTreeEdit from './FileTreeEdit.svelte'
  import FileTreeMenu from './FileTreeMenu.svelte'
  import FileTreeNode from './FileTreeNode.svelte'

  let { node, depth = 0 }: { node: FileNode; depth?: number } = $props()

  const open = $derived(!$collapsedFolders.has(node.path))
  const indent = $derived(`${14 + depth * 12}px`)
  const hasProblems = $derived(!node.isDir && $filesWithProblems.includes(node.path))
  const renaming = $derived($treeEdit?.kind === 'rename' && $treeEdit.path === node.path)
  const creatingHere = $derived(
    node.isDir && $treeEdit !== null && $treeEdit.kind !== 'rename' && $treeEdit.path === node.path
  )
  const showChildren = $derived(open && (node.children.length > 0 || creatingHere))

  const onKeydown = (event: KeyboardEvent): void => {
    if (event.key === 'F2' && depth > 0) {
      event.preventDefault()
      startRename(node.path)
    } else if (event.key === 'Delete' && depth > 0) {
      event.preventDefault()
      void deleteEntry(bridge, node.path)
    } else if (event.key === 'Enter' && !node.isDir) {
      event.preventDefault()
      void openFile(bridge, node.path)
    }
  }
</script>

{#snippet dirRow(props: Record<string, unknown>)}
  <button
    {...props}
    type="button"
    class="file dir"
    class:sel={node.path === $selectedNode}
    style:padding-left={indent}
    aria-expanded={open}
    aria-label={$t(open ? 'panels.collapseFolder' : 'panels.expandFolder', {
      values: { name: node.name }
    })}
    onclick={() => {
      selectedNode.set(node.path)
      toggleFolder(node.path)
    }}
    onkeydown={onKeydown}
  >
    <span aria-hidden="true">{open ? '▾' : '▸'}</span>
    {node.name}
  </button>
{/snippet}

{#snippet fileRow(props: Record<string, unknown>)}
  <button
    {...props}
    type="button"
    class="file"
    data-path={node.path}
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
{/snippet}

<li>
  {#if renaming && $treeEdit}
    <FileTreeEdit
      initial={node.name}
      {indent}
      error={$treeEdit.error}
      oncommit={(value) => void commitEdit(bridge, value)}
      oncancel={cancelEdit}
      oninput={clearEditError}
    />
  {:else}
    <FileTreeMenu path={node.path} row={node.isDir ? dirRow : fileRow} />
  {/if}
  {#if node.isDir && showChildren}
    <ul>
      {#if creatingHere && $treeEdit}
        <li>
          <FileTreeEdit
            indent={`${14 + (depth + 1) * 12}px`}
            error={$treeEdit.error}
            oncommit={(value) => void commitEdit(bridge, value)}
            oncancel={cancelEdit}
            oninput={clearEditError}
          />
        </li>
      {/if}
      {#each node.children as child (child.path)}
        <FileTreeNode node={child} depth={depth + 1} />
      {/each}
    </ul>
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
