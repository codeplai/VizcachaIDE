<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { copyText } from '../stores'

  /** A command the student pastes in a terminal: shown in full, with a button to copy it. */
  let { command }: { command: string } = $props()
</script>

<div class="command">
  <code>{command}</code>
  <button type="button" class="dlg-button" onclick={() => copyText(bridge, command)}>
    {$t('errors.toolCopyCommand')}
  </button>
</div>

<style>
  .command {
    display: grid;
    gap: 8px;
    justify-items: start;
  }
  code {
    font: 400 13px var(--mono);
    padding: 6px 8px;
    border-radius: 6px;
    background: var(--rail);
    user-select: all;
    overflow-wrap: anywhere;
  }
</style>
