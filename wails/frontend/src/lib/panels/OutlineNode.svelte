<script lang="ts">
  import { bridge } from '../bridge'
  import type { DocumentSymbol } from '../domain'
  import { goToLocation, symbolGlyph } from '../stores'
  import OutlineNode from './OutlineNode.svelte'

  let { symbol, depth = 0 }: { symbol: DocumentSymbol; depth?: number } = $props()
</script>

<li>
  <button
    type="button"
    class="symbol"
    style:padding-left={`${14 + depth * 14}px`}
    onclick={() => goToLocation(bridge, symbol.location)}
  >
    <span class="kind" aria-hidden="true">{symbolGlyph(symbol.kind)}</span>
    <span class="name">{symbol.name}</span>
    <span class="line">{symbol.location.line}</span>
  </button>
  {#if symbol.children.length > 0}
    <ul>
      {#each symbol.children as child (child.name + child.location.line)}
        <OutlineNode symbol={child} depth={depth + 1} />
      {/each}
    </ul>
  {/if}
</li>

<style>
  li {
    list-style: none;
  }
  ul {
    margin: 0;
    padding: 0;
  }
  .symbol {
    width: 100%;
    border: 0;
    background: none;
    cursor: pointer;
    display: flex;
    gap: 8px;
    align-items: baseline;
    padding: 4px 14px;
    font: 400 13px var(--mono);
    text-align: left;
  }
  .symbol:hover {
    background: var(--rail);
  }
  .kind {
    color: var(--code-fn);
    width: 1ch;
  }
  .line {
    margin-left: auto;
    color: var(--muted);
    font-size: 11px;
  }
</style>
