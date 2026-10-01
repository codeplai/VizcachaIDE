<script lang="ts">
  import { bridge } from '../bridge'
  import type { Settings } from '../domain'
  import { t } from '../i18n'
  import { settings, toolOrigin, toolchain, updateSettings } from '../stores'

  type PathKey = 'goPath' | 'delvePath' | 'goplsPath'
  type VersionKey = 'goVersion' | 'delveVersion' | 'goplsVersion'

  const tools: { label: string; path: PathKey; version: VersionKey }[] = [
    { label: 'settings.toolGo', path: 'goPath', version: 'goVersion' },
    { label: 'settings.toolDelve', path: 'delvePath', version: 'delveVersion' },
    { label: 'settings.toolGopls', path: 'goplsPath', version: 'goplsVersion' }
  ]

  const pathOf = (current: Settings | null, key: PathKey): string => current?.[key] ?? ''
</script>

{#each tools as tool (tool.path)}
  {@const path = pathOf($settings, tool.path)}
  {@const version = $toolchain?.[tool.version]}
  <div class="field">
    <label for={`setting-${tool.path}`}>{$t(tool.label)}</label>
    <input
      id={`setting-${tool.path}`}
      type="text"
      value={path}
      placeholder={$t('settings.pathPlaceholder')}
      aria-describedby={`setting-${tool.path}-origin`}
      onchange={(event) =>
        updateSettings(bridge, { [tool.path]: event.currentTarget.value.trim() })}
    />
    <span class="hint" id={`setting-${tool.path}-origin`}>
      {$t(toolOrigin(path) === 'custom' ? 'settings.originCustom' : 'settings.originAutomatic')}
      ·
      {version ? $t('settings.version', { values: { version } }) : $t('settings.notFound')}
    </span>
  </div>
{/each}
