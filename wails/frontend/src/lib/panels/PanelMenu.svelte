<script lang="ts" module>
  /** One entry of a text panel's right-click menu. `selection` is the text selected when it opened. */
  export interface PanelAction {
    label: string
    run: (selection: string) => void
    keys?: string
    disabled?: boolean
    /** Draws a separator line before this entry. */
    separated?: boolean
  }
</script>

<script lang="ts">
  import { ContextMenu } from 'bits-ui'
  import type { Snippet } from 'svelte'
  import { t } from '../i18n'
  import { selectedText } from '../stores'

  interface Props {
    actions: (selection: string) => PanelAction[]
    children: Snippet
  }
  let { actions, children }: Props = $props()

  // Read the selection when the menu opens: clicking an entry may clear it.
  let selection = $state('')
</script>

<ContextMenu.Root onOpenChange={(open) => open && (selection = selectedText())}>
  <ContextMenu.Trigger class="panel-menu-area">
    {@render children()}
  </ContextMenu.Trigger>
  <ContextMenu.Portal>
    <ContextMenu.Content class="menu">
      {#each actions(selection) as action (action.label)}
        {#if action.separated}<ContextMenu.Separator class="menu-sep" />{/if}
        <ContextMenu.Item
          class="menu-item panel-menu-row"
          disabled={action.disabled}
          onSelect={() => action.run(selection)}
        >
          {$t(action.label)}
          {#if action.keys}<kbd>{action.keys}</kbd>{/if}
        </ContextMenu.Item>
      {/each}
    </ContextMenu.Content>
  </ContextMenu.Portal>
</ContextMenu.Root>

<style>
  :global(.panel-menu-area) {
    display: contents;
  }
  :global(.menu-item.panel-menu-row) {
    display: flex;
    justify-content: space-between;
    gap: 24px;
  }
  :global(.panel-menu-row[data-disabled]) {
    opacity: 0.45;
    cursor: default;
  }
  :global(.panel-menu-row kbd) {
    font: 400 12px var(--mono);
    color: var(--muted);
  }
</style>
