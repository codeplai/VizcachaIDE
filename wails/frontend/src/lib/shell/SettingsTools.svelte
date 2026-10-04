<script lang="ts">
  import { bridge, type ToolId } from '../bridge'
  import type { Settings } from '../domain'
  import { t } from '../i18n'
  import { pickTool, settings, toolSourceKey, toolchain, updateSettings } from '../stores'

  type VersionKey = 'goVersion' | 'delveVersion' | 'goplsVersion'
  type SourceKey = 'goSource' | 'delveSource' | 'goplsSource'

  interface Tool {
    id: ToolId
    label: string
    version: VersionKey
    source: SourceKey
  }

  const tools: Tool[] = [
    {
      id: 'go',
      label: 'settings.toolGo',
      version: 'goVersion',
      source: 'goSource'
    },
    {
      id: 'dlv',
      label: 'settings.toolDelve',
      version: 'delveVersion',
      source: 'delveSource'
    },
    {
      id: 'gopls',
      label: 'settings.toolGopls',
      version: 'goplsVersion',
      source: 'goplsSource'
    }
  ]

  const pathOf = (current: Settings | null, id: ToolId): string => current?.toolPaths[id] ?? ''

  const savePath = (id: ToolId, value: string): Promise<void> =>
    updateSettings(bridge, { toolPaths: { ...($settings?.toolPaths ?? {}), [id]: value } })
</script>

{#each tools as tool (tool.id)}
  {@const path = pathOf($settings, tool.id)}
  {@const version = $toolchain?.[tool.version]}
  {@const source = $toolchain?.[tool.source]}
  <div class="field">
    <label for={`setting-${tool.id}`}>{$t(tool.label)}</label>
    <div class="row">
      <input
        id={`setting-${tool.id}`}
        type="text"
        value={path}
        placeholder={$t('settings.pathPlaceholder')}
        aria-describedby={`setting-${tool.id}-origin`}
        onchange={(event) => savePath(tool.id, event.currentTarget.value.trim())}
      />
      <button type="button" class="dlg-button plain" onclick={() => pickTool(bridge, tool.id)}>
        {$t('settings.chooseTool')}
      </button>
    </div>
    <span class="hint" id={`setting-${tool.id}-origin`}>
      {#if source}{$t(toolSourceKey(source))} ·{/if}
      {version ? $t('settings.version', { values: { version } }) : $t('settings.notFound')}
    </span>
  </div>
{/each}

<style>
  .row {
    display: flex;
    gap: 8px;
  }
  .row input {
    flex: 1;
    min-width: 0;
  }
</style>
