<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { openDialog, toolSourceKey, tools } from '../stores'
  import logo from '../../assets/brand/logo.png'
  import Modal from './Modal.svelte'

  /** Product names are not translated; a tool without an entry shows its id. */
  const TOOL_NAMES: Record<string, string> = {
    go: 'Go',
    dlv: 'Delve',
    gopls: 'gopls',
    python: 'CPython',
    debugpy: 'debugpy',
    pylsp: 'python-lsp-server',
    ruff: 'ruff',
    cxx: 'g++ / clang++',
    'lldb-dap': 'lldb-dap',
    clangd: 'clangd',
    'clang-format': 'clang-format',
    rustc: 'rustc',
    cargo: 'Cargo',
    'rust-analyzer': 'rust-analyzer',
    clippy: 'clippy',
    rustfmt: 'rustfmt'
  }
  /** Third-party projects the C++ and Rust toolchains come from (product names are not translated). */
  const toolchainCredits = [
    'LLVM (clang, lldb, clangd, clang-format) · MinGW-w64 · GCC',
    'Rust (rustc, cargo, clippy, rustfmt, rust-analyzer)'
  ]
  // Two languages can share a tool (lldb-dap): it is listed once.
  const listedTools = $derived(
    $tools.filter((tool, index) => $tools.findIndex((other) => other.id === tool.id) === index)
  )
  const author = { name: 'Marks Calderon', role: 'CEO Codeplai', email: 'hola@codeplai.pe' }
  const website = 'https://vizcacha.codeplai.pe'
  const licenseUrl = 'https://opensource.org/licenses/MIT'
  const copyright = 'Copyright © 2025-2026 Marks Calderon - Codeplai Games'
</script>

<Modal
  open={$openDialog === 'about'}
  title={$t('shell.about')}
  description={$t('shell.aboutTagline')}
  onClose={() => openDialog.set(null)}
>
  <div class="intro">
    <img class="about-logo" src={logo} alt={$t('app.name')} />
    <p class="version">{$t('shell.aboutVersion', { values: { version: __APP_VERSION__ } })}</p>
    <button type="button" class="link" onclick={() => bridge.system.openUrl(website)}>
      vizcacha.codeplai.pe
    </button>
    <p class="made-in">{$t('shell.aboutMadeIn')}</p>
  </div>
  <div class="field">
    <span class="label">{$t('shell.aboutTools')}</span>
    <dl>
      {#each listedTools as tool (tool.id)}
        <dt>{TOOL_NAMES[tool.id] ?? tool.id}</dt>
        <dd>
          {tool.version || $t('shell.aboutNotFound')}
          <small>{$t(toolSourceKey(tool.source))}</small>
        </dd>
      {/each}
    </dl>
  </div>
  <div class="credits">
    <p>{$t('shell.aboutCreatedBy')} <strong>{author.name}</strong></p>
    <p class="role">{author.role}</p>
    <button
      type="button"
      class="link mono"
      onclick={() => bridge.system.openUrl(`mailto:${author.email}`)}
    >
      {author.email}
    </button>
  </div>
  <div class="legal">
    <p>{copyright}</p>
    <p>{$t('shell.aboutWarranty')}</p>
    <p>{$t('shell.aboutBundled')}</p>
    {#each toolchainCredits as credit (credit)}<p>{credit}</p>{/each}
    <button type="button" class="link" onclick={() => bridge.system.openUrl(licenseUrl)}>
      {$t('shell.aboutLicenseLink')}
    </button>
  </div>
</Modal>

<style>
  .intro {
    display: grid;
    justify-items: center;
    gap: 2px;
    text-align: center;
  }
  .about-logo {
    display: block;
    width: 180px;
    max-width: 100%;
    height: auto;
    margin: 0 auto 4px;
  }
  .intro p {
    margin: 0;
    font-size: 14px;
  }
  .version {
    font-family: var(--mono);
  }
  .made-in {
    color: var(--muted);
  }
  dl {
    display: grid;
    grid-template-columns: max-content 1fr;
    gap: 4px 16px;
    margin: 0;
    font-size: 14px;
  }
  small {
    display: block;
    font: 400 12px var(--ui);
  }
  .credits {
    display: grid;
    justify-items: center;
    gap: 2px;
    text-align: center;
    font-size: 14px;
  }
  .credits p {
    margin: 0;
  }
  .role {
    color: var(--muted);
  }
  .link {
    padding: 0;
    border: 0;
    background: none;
    font: 400 14px var(--ui);
    color: var(--go);
    text-decoration: underline;
    cursor: pointer;
    user-select: text;
  }
  .mono {
    font-family: var(--mono);
  }
  .legal {
    display: grid;
    justify-items: center;
    gap: 4px;
    padding-top: 10px;
    border-top: 1px solid var(--line);
    text-align: center;
    font-size: 12px;
    color: var(--muted);
  }
  .legal p {
    margin: 0;
  }
  .legal .link {
    font-size: 12px;
  }
  dd {
    margin: 0;
    font-family: var(--mono);
    color: var(--muted);
  }
</style>
