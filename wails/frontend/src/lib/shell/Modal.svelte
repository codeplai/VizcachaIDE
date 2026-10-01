<script lang="ts">
  import { Dialog } from 'bits-ui'
  import type { Snippet } from 'svelte'
  import { t } from '../i18n'

  interface Props {
    open: boolean
    title: string
    onClose: () => void
    /** Optional second line under the title. */
    description?: string
    children: Snippet
    actions?: Snippet
  }

  let { open, title, onClose, description, children, actions }: Props = $props()
</script>

<Dialog.Root {open} onOpenChange={(next) => !next && onClose()}>
  <Dialog.Portal>
    <Dialog.Overlay class="dlg-overlay" />
    <Dialog.Content class="dlg" interactOutsideBehavior="close">
      <Dialog.Title>{title}</Dialog.Title>
      {#if description}
        <Dialog.Description>{description}</Dialog.Description>
      {:else}
        <Dialog.Description class="sr-only">{title}</Dialog.Description>
      {/if}
      {@render children()}
      <div class="dlg-actions">
        {#if actions}
          {@render actions()}
        {:else}
          <Dialog.Close class="dlg-button primary">{$t('shell.close')}</Dialog.Close>
        {/if}
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<style>
  :global(.sr-only) {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
</style>
