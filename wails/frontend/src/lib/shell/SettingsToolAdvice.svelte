<script lang="ts">
  import { t } from '../i18n'
  import { RUST_ADVICE_COMMANDS, rustAdvice } from '../stores'
  import CopyCommand from './CopyCommand.svelte'

  /** The non-blocking advice about the Rust install, each with the command that fixes it. */
  const items = $derived($rustAdvice.filter((key) => key in RUST_ADVICE_COMMANDS))
</script>

{#each items as key (key)}
  <div class="field advice" role="note" data-advice={key}>
    <span class="hint">{$t(key)}</span>
    <CopyCommand command={RUST_ADVICE_COMMANDS[key] ?? ''} />
  </div>
{/each}
