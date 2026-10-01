<script lang="ts">
  import { t } from '../i18n'
  import { openDialog, toolchain } from '../stores'
  import Modal from './Modal.svelte'

  const tools = [
    { name: 'Go', key: 'goVersion' },
    { name: 'Delve', key: 'delveVersion' },
    { name: 'gopls', key: 'goplsVersion' }
  ] as const
</script>

<Modal
  open={$openDialog === 'about'}
  title={$t('shell.about')}
  description={$t('shell.aboutTagline')}
  onClose={() => openDialog.set(null)}
>
  <div class="field">
    <span class="label">{$t('shell.aboutTools')}</span>
    <dl>
      {#each tools as tool (tool.name)}
        <dt>{tool.name}</dt>
        <dd>{$toolchain?.[tool.key] || $t('shell.aboutNotFound')}</dd>
      {/each}
    </dl>
  </div>
</Modal>

<style>
  dl {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 4px 16px;
    margin: 0;
    font-size: 14px;
  }
  dd {
    margin: 0;
    font-family: var(--mono);
    color: var(--muted);
  }
</style>
