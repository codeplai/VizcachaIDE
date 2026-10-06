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
    dropTarget,
    endDrag,
    filesWithProblems,
    openFile,
    selectedPaths,
    treeEdit
  } from '../stores'
  import { rowEvents } from './fileTreeRow'
  import FileTreeEdit from './FileTreeEdit.svelte'
  import FileTreeMenu from './FileTreeMenu.svelte'
  import FileTreeNode from './FileTreeNode.svelte'

  const GENERATED_FOLDERS = ['target', 'build']

  let { node, depth = 0 }: { node: FileNode; depth?: number } = $props()

  const open = $derived(!$collapsedFolders.has(node.path))
  const indent = $derived(`${14 + depth * 12}px`)
  // Folders the build tools generate (Cargo's target/, CMake's build/) are shown dimmed.
  const generated = $derived(node.isDir && depth > 0 && GENERATED_FOLDERS.includes(node.name))
  const hasProblems = $derived(!node.isDir && $filesWithProblems.includes(node.path))
  const renaming = $derived($treeEdit?.kind === 'rename' && $treeEdit.path === node.path)
  const creatingHere = $derived(
    node.isDir && $treeEdit !== null && $treeEdit.kind !== 'rename' && $treeEdit.path === node.path
  )
  const showChildren = $derived(open && (node.children.length > 0 || creatingHere))

  const selected = $derived($selectedPaths.includes(node.path))
  const dropping = $derived(node.isDir && $dropTarget === node.path)

  const events = $derived(rowEvents(bridge, node, depth))
</script>

{#snippet dirRow(props: Record<string, unknown>)}
  <button
    {...props}
    type="button"
    class="file dir"
    class:sel={selected}
    class:drop={dropping}
    data-path={node.path}
    style:padding-left={indent}
    aria-expanded={open}
    aria-label={$t(open ? 'panels.collapseFolder' : 'panels.expandFolder', {
      values: { name: node.name }
    })}
    draggable={depth > 0}
    ondragstart={events.onDragStart}
    ondragend={endDrag}
    ondragover={events.onDragOver}
    ondragleave={() => dropping && dropTarget.set(null)}
    ondrop={events.onDrop}
    onclick={events.onClick}
    onkeydown={events.onKeydown}
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
    class:sel={selected}
    style:padding-left={indent}
    aria-current={node.path === $activePath ? 'true' : undefined}
    draggable={depth > 0}
    ondragstart={events.onDragStart}
    ondragend={endDrag}
    ondragover={events.onDragOver}
    ondragleave={() => dropping && dropTarget.set(null)}
    ondrop={events.onDrop}
    onclick={events.onClick}
    ondblclick={() => openFile(bridge, node.path)}
    onkeydown={events.onKeydown}
  >
    {node.name}
    {#if hasProblems}
      <span class="dot" role="img" aria-label={$t('panels.hasProblems')}></span>
    {/if}
  </button>
{/snippet}

<li class:generated>
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
  li.generated {
    opacity: 0.55;
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
  .file.drop {
    background: var(--go-soft);
    outline: 1px dashed var(--go);
    outline-offset: -1px;
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
