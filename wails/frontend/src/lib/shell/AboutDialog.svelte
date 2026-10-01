<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { openDialog, toolSourceKey, toolchain } from '../stores'
  import logo from '../../assets/brand/logo.png'
  import Modal from './Modal.svelte'

  const tools = [
    { name: 'Go', key: 'goVersion', source: 'goSource' },
    { name: 'Delve', key: 'delveVersion', source: 'delveSource' },
    { name: 'gopls', key: 'goplsVersion', source: 'goplsSource' }
  ] as const
  const author = { name: 'Marks Calderon', role: 'CEO Codeplai', email: 'hola@codeplai.pe' }
</script>

<Modal
  open={$openDialog === 'about'}
  title={$t('shell.about')}
  description={$t('shell.aboutTagline')}
  onClose={() => openDialog.set(null)}
>
  <img class="about-logo" src={logo} alt={$t('app.name')} />
  <div class="field">
    <span class="label">{$t('shell.aboutTools')}</span>
    <dl>
      {#each tools as tool (tool.name)}
        <dt>{tool.name}</dt>
        <dd>
          {$toolchain?.[tool.key] || $t('shell.aboutNotFound')}
          {#if $toolchain}<small>{$t(toolSourceKey($toolchain[tool.source]))}</small>{/if}
        </dd>
      {/each}
    </dl>
  </div>
  <div class="field">
    <span class="label">{$t('shell.aboutCreatedBy')}</span>
    <p class="author"><strong>{author.name}</strong><span>{author.role}</span></p>
  </div>
  <div class="field">
    <span class="label">{$t('shell.aboutContact')}</span>
    <button
      type="button"
      class="email"
      onclick={() => bridge.system.openUrl(`mailto:${author.email}`)}
    >
      {author.email}
    </button>
  </div>
</Modal>

<style>
  .about-logo {
    display: block;
    width: 220px;
    max-width: 100%;
    height: auto;
    margin: 0 auto 8px;
  }
  dl {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 4px 16px;
    margin: 0;
    font-size: 14px;
  }
  small {
    display: block;
    font: 400 12px var(--ui);
  }
  .author {
    display: grid;
    margin: 0;
    font-size: 14px;
  }
  .author span {
    color: var(--muted);
  }
  .email {
    padding: 0;
    border: 0;
    background: none;
    font: 400 14px var(--mono);
    color: var(--go);
    text-decoration: underline;
    cursor: pointer;
    user-select: text;
  }
  dd {
    margin: 0;
    font-family: var(--mono);
    color: var(--muted);
  }
</style>
