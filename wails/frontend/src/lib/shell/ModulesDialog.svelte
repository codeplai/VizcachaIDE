<script lang="ts">
  import { t } from '../i18n'
  import { lastRunConfiguration, openDialog } from '../stores'
  import Modal from './Modal.svelte'

  const module = $derived($lastRunConfiguration?.module ?? null)
</script>

<Modal
  open={$openDialog === 'modules'}
  title={$t('shell.goModules')}
  description={$t('shell.modulesIntro')}
  onClose={() => openDialog.set(null)}
>
  {#if module}
    <dl>
      <dt>{$t('shell.modulesName')}</dt>
      <dd>{module.modulePath}</dd>
      <dt>{$t('shell.modulesFolder')}</dt>
      <dd>{module.root}</dd>
    </dl>
  {:else}
    <p class="none">{$t('shell.modulesNone')}</p>
  {/if}
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
    overflow-wrap: anywhere;
  }
</style>
