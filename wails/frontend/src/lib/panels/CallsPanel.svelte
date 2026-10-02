<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { argumentsText, calls, frameDetails, goToLocation, loadFrameDetails } from '../stores'

  $effect(() => {
    for (const box of $calls.boxes) void loadFrameDetails(bridge, box.frameId)
  })
</script>

{#snippet box(index: number)}
  {@const item = $calls.boxes[index]}
  {#if item}
    {@const details = $frameDetails[item.frameId]}
    {@const location = item.location}
    <div class="call" class:current={item.current} data-testid="call-box">
      <button
        type="button"
        class="head"
        disabled={!location}
        aria-current={item.current ? 'true' : undefined}
        aria-label={location
          ? $t('panels.callsGoTo', { values: { line: location.line, function: item.name } })
          : item.name}
        onclick={() => location && goToLocation(bridge, location)}
      >
        <span class="title">{item.name}({argumentsText(details)})</span>
        {#if location}<span class="line">:{location.line}</span>{/if}
        {#if item.current}<span class="here">{$t('panels.callsCurrent')}</span>{/if}
      </button>
      {#if details && details.locals.length > 0}
        <ul class="vars">
          {#each details.locals as variable, position (position)}
            <li>
              <span class="n">{variable.name}</span>
              <span class="v">{variable.value}</span>
            </li>
          {/each}
        </ul>
      {/if}
      {#if index + 1 < $calls.boxes.length}{@render box(index + 1)}{/if}
    </div>
  {/if}
{/snippet}

<div class="calls">
  <p class="hint">{$t('panels.callsHint')}</p>
  {#if $calls.boxes.length === 0}
    <p class="empty">{$t('panels.callsEmpty')}</p>
  {:else}
    {#if $calls.hidden > 0}
      <p class="more">{$t('panels.callsMore', { values: { count: $calls.hidden } })}</p>
    {/if}
    {@render box(0)}
  {/if}
</div>

<style>
  .calls {
    display: grid;
    gap: 8px;
    min-width: 0;
  }
  .hint,
  .empty,
  .more {
    margin: 0;
    color: var(--muted);
    font-size: 12.5px;
  }
  .more {
    font: 600 12px var(--mono);
  }
  .call {
    display: grid;
    gap: 4px;
    padding: 5px 4px 5px 6px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--win);
    min-width: 0;
  }
  .call.current {
    border-color: var(--sand);
    background: var(--sand-soft);
    box-shadow: inset 3px 0 0 var(--sand);
  }
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 2px 6px;
    border: 0;
    background: none;
    padding: 2px 2px;
    border-radius: 6px;
    text-align: left;
    font: 600 12.5px var(--mono);
    color: var(--ink);
    cursor: pointer;
    overflow-wrap: anywhere;
  }
  .head:hover:not(:disabled) {
    background: var(--rail);
  }
  .head:disabled {
    cursor: default;
  }
  .line {
    font-weight: 400;
    color: var(--muted);
  }
  .here {
    margin-left: auto;
    font: 600 11px var(--ui);
    color: var(--ink);
    background: var(--sand);
    border-radius: 999px;
    padding: 0 8px;
  }
  .vars {
    display: grid;
    gap: 2px;
    margin: 0;
    padding: 0 2px;
  }
  .vars li {
    list-style: none;
    display: flex;
    justify-content: space-between;
    gap: 8px;
    font: 400 12.5px var(--mono);
    min-width: 0;
  }
  .n {
    color: var(--muted);
  }
  .v {
    color: var(--go);
    font-weight: 600;
    overflow-wrap: anywhere;
    text-align: right;
  }
</style>
