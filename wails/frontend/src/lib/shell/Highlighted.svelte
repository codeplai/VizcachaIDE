<script lang="ts">
  import { highlightSegments } from '../search/highlight'

  interface Props {
    text: string
    /** Positions that matched, counted from `offset` before the start of `text`. */
    indices: number[]
    offset?: number
  }

  let { text, indices, offset = 0 }: Props = $props()
</script>

{#each highlightSegments(text, indices, offset) as part, at (at)}{#if part.matched}<mark
      >{part.text}</mark
    >{:else}{part.text}{/if}{/each}

<style>
  mark {
    background: none;
    color: var(--go);
    font-weight: 800;
    text-decoration: underline;
  }
</style>
