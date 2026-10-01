<script lang="ts">
  import { bridge, type ToolId } from '../bridge'
  import type { Settings } from '../domain'
  import { t } from '../i18n'
  import { pickTool, settings, toolSourceKey, toolchain, updateSettings } from '../stores'

  type PathKey = 'goPath' | 'delvePath' | 'goplsPath'
  type VersionKey = 'goVersion' | 'delveVersion' | 'goplsVersion'
  type SourceKey = 'goSource' | 'delveSource' | 'goplsSource'

  interface Tool {
    id: ToolId
    label: string
    path: PathKey
    version: VersionKey
    source: SourceKey
  }

  const tools: Tool[] = [
    {
      id: 'go',
      label: 'settings.toolGo',
      path: 'goPath',
      version: 'goVersion',
      source: 'goSource'
    },
    {
      id: 'dlv',
      label: 'settings.toolDelve',
      path: 'delvePath',
      version: 'delveVersion',
      source: 'delveSource'
    },
    {
      id: 'gopls',
      label: 'settings.toolGopls',
      path: 'goplsPath',
      version: 'goplsVersion',
      source: 'goplsSource'
    }
  ]

  const pathOf = (current: Settings | null, key: PathKey): string => current?.[key] ?? ''
</script>

{#each tools as tool (tool.path)}
  {@const path = pathOf($settings, tool.path)}
  {@const version = $toolchain?.[tool.version]}
  {@const source = $toolchain?.[tool.source]}
  <div class="field">
    <label for={`setting-${tool.path}`}>{$t(tool.label)}</label>
    <div class="row">
      <input
        id={`setting-${tool.path}`}
        type="text"
        value={path}
        placeholder={$t('settings.pathPlaceholder')}
        aria-describedby={`setting-${tool.path}-origin`}
        onchange={(event) =>
          updateSettings(bridge, { [tool.path]: event.currentTarget.value.trim() })}
      />
      <button type="button" class="dlg-button plain" onclick={() => pickTool(bridge, tool.id)}>
        {$t('settings.chooseTool')}
      </button>
    </div>
    <span class="hint" id={`setting-${tool.path}-origin`}>
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
