<script lang="ts">
  import { bridge } from '../bridge'
  import type { ToolSpec } from '../domain'
  import { t } from '../i18n'
  import { detectPlatform, type SystemPlatform } from '../platform'
  import { copyText } from '../stores'

  /** The C++ compiler's spec: its install link is the Windows one. */
  let { spec, platform = detectPlatform() }: { spec: ToolSpec; platform?: SystemPlatform } =
    $props()

  const MACOS_COMMAND = 'xcode-select --install'
  const LINUX_COMMAND = 'sudo apt install g++ lldb clangd clang-format'
  const command = $derived(platform === 'macos' ? MACOS_COMMAND : LINUX_COMMAND)
</script>

<div class="hint" data-platform={platform}>
  {#if platform === 'windows'}
    <p>{$t('firstRun.cppWindows')}</p>
    {#if spec.installUrl}
      <button
        type="button"
        class="dlg-button"
        onclick={() => bridge.system.openUrl(spec.installUrl)}
      >
        {$t('errors.toolInstall')}
      </button>
    {/if}
  {:else}
    {#if platform === 'macos'}<p>{$t('errors.cppXcodeTools')}</p>{/if}
    <code>{command}</code>
    <button type="button" class="dlg-button" onclick={() => copyText(bridge, command)}>
      {$t('errors.toolCopyCommand')}
    </button>
  {/if}
</div>

<style>
  .hint {
    display: grid;
    gap: 8px;
    justify-items: start;
  }
  p {
    margin: 0;
  }
  code {
    font: 400 13px var(--mono);
    padding: 6px 8px;
    border-radius: 6px;
    background: var(--rail);
    user-select: all;
  }
</style>
