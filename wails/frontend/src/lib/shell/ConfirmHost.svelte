<script lang="ts">
  import { AlertDialog } from 'bits-ui'
  import { t } from '../i18n'
  import { pendingConfirm } from '../stores'
</script>

<!-- One place that shows every confirmation (close with changes, reload, replace all). -->
<AlertDialog.Root
  open={$pendingConfirm !== null}
  onOpenChange={(next) => !next && $pendingConfirm?.answer($pendingConfirm.dismissId)}
>
  <AlertDialog.Portal>
    <AlertDialog.Overlay class="dlg-overlay" />
    <AlertDialog.Content class="dlg">
      {#if $pendingConfirm}
        {@const request = $pendingConfirm}
        <AlertDialog.Title class="sr-only">{$t('app.name')}</AlertDialog.Title>
        <AlertDialog.Description class="confirm-message">
          {$t(request.messageKey, { values: request.values })}
        </AlertDialog.Description>
        <div class="dlg-actions">
          {#each request.choices as choice (choice.id)}
            <button
              type="button"
              class={`dlg-button ${choice.tone}`}
              onclick={() => request.answer(choice.id)}
            >
              {$t(choice.labelKey, { values: choice.values })}
            </button>
          {/each}
        </div>
      {/if}
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>

<style>
  :global(.confirm-message) {
    color: var(--ink) !important;
    font-size: 15px !important;
  }
</style>
