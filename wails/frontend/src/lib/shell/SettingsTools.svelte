<script lang="ts">
  import { bridge, type ToolId } from '../bridge'
  import type { Settings, ToolSpec } from '../domain'
  import { t } from '../i18n'
  import {
    pickTool,
    profiles,
    settings,
    toolSourceKey,
    toolStatus,
    tools,
    updateSettings
  } from '../stores'

  const pathOf = (current: Settings | null, id: ToolId): string => current?.toolPaths[id] ?? ''

  const savePath = (id: ToolId, value: string): Promise<void> =>
    updateSettings(bridge, { toolPaths: { ...($settings?.toolPaths ?? {}), [id]: value } })

  /** A tool that lives inside another one (debugpy in python) has no path of its own. */
  const hasOwnPath = (spec: ToolSpec): boolean => spec.providedBy === ''
</script>

{#each $profiles as profile (profile.id)}
  <h3 class="group">{$t('settings.toolsOf', { values: { codeLanguage: $t(profile.nameKey) } })}</h3>
  {#each profile.tools as spec (spec.id)}
    {@const status = toolStatus(spec.id, $tools)}
    {@const version = status?.version}
    <div class="field">
      {#if hasOwnPath(spec)}
        <label for={`setting-${spec.id}`}>{$t(spec.labelKey)}</label>
        <div class="row">
          <input
            id={`setting-${spec.id}`}
            type="text"
            value={pathOf($settings, spec.id)}
            placeholder={$t('settings.pathPlaceholder')}
            aria-describedby={`setting-${spec.id}-origin`}
            onchange={(event) => savePath(spec.id, event.currentTarget.value.trim())}
          />
          <button type="button" class="dlg-button plain" onclick={() => pickTool(bridge, spec.id)}>
            {$t('settings.chooseTool')}
          </button>
        </div>
      {:else}
        <span class="label">{$t(spec.labelKey)}</span>
      {/if}
      <span class="hint" id={`setting-${spec.id}-origin`}>
        {#if status}{$t(toolSourceKey(status.source))} ·{/if}
        {version ? $t('settings.version', { values: { version } }) : $t('settings.notFound')}
      </span>
    </div>
  {/each}
{/each}

<style>
  .group {
    margin: 4px 0 0;
    font-size: 13px;
    font-weight: 700;
    color: var(--muted);
  }
  .row {
    display: flex;
    gap: 8px;
  }
  .row input {
    flex: 1;
    min-width: 0;
  }
</style>
