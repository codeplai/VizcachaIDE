<script lang="ts">
  import { onMount } from 'svelte'
  import { t } from '../i18n'

  interface Props {
    initial?: string
    indent: string
    error: { key: string; values: Record<string, string> } | null
    oncommit: (value: string) => void
    oncancel: () => void
    oninput: () => void
  }
  let { initial = '', indent, error, oncommit, oncancel, oninput }: Props = $props()

  let input: HTMLInputElement | undefined = $state()
  // The box starts with the name it was opened with; later changes of `initial` do not matter.
  // svelte-ignore state_referenced_locally
  let value = $state(initial)
  /** Enter or Esc already decided: the blur that follows must not decide again. */
  let decided = false

  onMount(() => {
    input?.focus()
    // Select the name without the extension, like a file manager does.
    const dot = initial.lastIndexOf('.')
    input?.setSelectionRange(0, dot > 0 ? dot : initial.length)
  })

  const onKeydown = (event: KeyboardEvent): void => {
    event.stopPropagation()
    if (event.key === 'Enter') {
      event.preventDefault()
      oncommit(value)
    } else if (event.key === 'Escape') {
      event.preventDefault()
      decided = true
      oncancel()
    }
  }

  const onBlur = (): void => {
    if (decided) return
    decided = true
    oncancel()
  }
</script>

<div class="edit" style:padding-left={indent}>
  <input
    bind:this={input}
    bind:value
    type="text"
    class="name"
    class:bad={error !== null}
    spellcheck="false"
    autocomplete="off"
    aria-label={$t('tree.nameLabel')}
    aria-invalid={error !== null}
    aria-describedby={error ? 'tree-edit-error' : undefined}
    onkeydown={onKeydown}
    oninput={() => oninput()}
    onblur={onBlur}
  />
  {#if error}
    <p class="error" id="tree-edit-error" role="alert">{$t(error.key, { values: error.values })}</p>
  {/if}
</div>

<style>
  .edit {
    padding-right: 10px;
    padding-top: 2px;
    padding-bottom: 2px;
  }
  .name {
    width: 100%;
    box-sizing: border-box;
    font: 13px var(--ui);
    color: var(--ink);
    background: var(--win);
    border: 1px solid var(--go);
    border-radius: 5px;
    padding: 3px 6px;
  }
  .name.bad {
    border-color: var(--err);
  }
  .error {
    margin: 3px 0 0;
    font-size: 12px;
    color: var(--err);
  }
</style>
