<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { fileTree, openFolder } from '../stores'
  import FileTreeNode from './FileTreeNode.svelte'
</script>

<section class="side">
  <h2 class="side-h">{$t('panels.files')}</h2>
  {#if $fileTree}
    <ul class="tree" aria-label={$t('a11y.fileTree')}>
      <FileTreeNode node={$fileTree} />
    </ul>
  {:else}
    <p class="empty">{$t('empty.files')}</p>
    <button type="button" class="mini cta" onclick={() => openFolder(bridge)}>
      {$t('empty.filesAction')}
    </button>
  {/if}
</section>

<style>
  .tree {
    margin: 0;
    padding: 0;
  }
  .cta {
    justify-self: start;
    margin: 6px 14px;
  }
</style>
