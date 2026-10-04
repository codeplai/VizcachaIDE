<script lang="ts">
  import { bridge, mockControls, type Scenario } from '../bridge'
  import { demoBreakpointLine, parseDevQuery } from '../devQuery'
  import { t } from '../i18n'
  import {
    activePath,
    breakpoints,
    debuggedPath,
    settings,
    toggleBreakpoint,
    updateSettings
  } from '../stores'
  import { get } from 'svelte/store'

  const states: { scenario: Scenario; label: string }[] = [
    { scenario: 'write', label: 'dev.stateWrite' },
    { scenario: 'error', label: 'dev.stateError' },
    { scenario: 'debug', label: 'dev.stateDebug' }
  ]
  const languages = ['es', 'en'] as const
  const themes = [
    { value: 'system', label: 'dev.themeSystem' },
    { value: 'light', label: 'dev.themeLight' },
    { value: 'dark', label: 'dev.themeDark' }
  ] as const

  let current = $state<Scenario>(parseDevQuery(location.search).scenario)

  const play = async (scenario: Scenario): Promise<void> => {
    current = scenario
    const path = get(activePath)
    const set = path ? (get(breakpoints)[path] ?? []) : []
    if (scenario === 'debug' && path) debuggedPath.set(path)
    const line = path ? demoBreakpointLine(path) : 0
    if (scenario === 'debug' && path && !set.includes(line))
      await toggleBreakpoint(bridge, path, line)
    await mockControls?.play(scenario)
  }
</script>

{#if mockControls}
  <div class="dev">
    <span class="badge">{$t('dev.mockBadge')}</span>
    <div class="seg" role="group" aria-label={$t('a11y.windowState')}>
      {#each states as state (state.scenario)}
        <button
          type="button"
          aria-pressed={current === state.scenario}
          onclick={() => play(state.scenario)}
        >
          {$t(state.label)}
        </button>
      {/each}
    </div>
    <div class="seg" role="group" aria-label={$t('a11y.language')}>
      {#each languages as language (language)}
        <button
          type="button"
          aria-pressed={$settings?.language === language}
          onclick={() => updateSettings(bridge, { language })}
        >
          {$t(`language.${language}`)}
        </button>
      {/each}
    </div>
    <div class="seg" role="group">
      {#each themes as theme (theme.value)}
        <button
          type="button"
          aria-pressed={$settings?.theme === theme.value}
          onclick={() => updateSettings(bridge, { theme: theme.value })}
        >
          {$t(theme.label)}
        </button>
      {/each}
    </div>
  </div>
{/if}

<style>
  .dev {
    flex: none;
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 4px 14px;
    background: var(--page);
    border-bottom: 1px solid var(--line);
  }
  .badge {
    font-size: 12px;
    color: var(--muted);
  }
  .seg {
    display: inline-flex;
    background: var(--win);
    border: 1px solid var(--line);
    border-radius: 10px;
    padding: 2px;
    gap: 2px;
  }
  .seg button {
    border: 0;
    background: none;
    color: var(--muted);
    font-weight: 600;
    font-size: 12px;
    padding: 4px 12px;
    border-radius: 7px;
    cursor: pointer;
  }
  .seg button[aria-pressed='true'] {
    background: var(--go);
    color: var(--win);
  }
</style>
