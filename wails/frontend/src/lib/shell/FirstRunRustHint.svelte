<script lang="ts">
  import { bridge } from '../bridge'
  import type { ToolSpec } from '../domain'
  import { t } from '../i18n'
  import { detectPlatform, type SystemPlatform } from '../platform'
  import { RUSTUP_URL, rustupInstallCommand } from '../stores'
  import CopyCommand from './CopyCommand.svelte'

  /** What to do when `rustc` is missing: rustup, with the GNU toolchain on Windows. */
  let { spec, platform = detectPlatform() }: { spec: ToolSpec; platform?: SystemPlatform } =
    $props()
</script>

<div class="hint" data-platform={platform}>
  <p>{$t('firstRun.rustInstall')}</p>
  <button
    type="button"
    class="dlg-button"
    onclick={() => bridge.system.openUrl(spec.installUrl || RUSTUP_URL)}
  >
    {$t('errors.toolInstall')} · rustup.rs
  </button>
  <CopyCommand command={rustupInstallCommand(platform)} />
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
</style>
