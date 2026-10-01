<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { sendProgramInput } from '../stores'

  let typed = $state('')

  const submit = (event: KeyboardEvent): void => {
    if (event.key !== 'Enter' || event.isComposing) return
    const text = typed
    typed = ''
    void sendProgramInput(bridge, text)
  }
</script>

<input
  class="stdin"
  type="text"
  spellcheck="false"
  autocomplete="off"
  placeholder={$t('run.inputPlaceholder')}
  aria-label={$t('run.inputPlaceholder')}
  bind:value={typed}
  onkeydown={submit}
/>

<style>
  .stdin {
    flex: none;
    margin: 0 14px 8px;
    padding: 6px 10px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--win);
    color: var(--ink);
    font: 400 13px var(--mono);
  }
</style>
