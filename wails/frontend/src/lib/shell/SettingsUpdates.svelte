<script lang="ts">
  import { bridge } from '../bridge'
  import { locale, t } from '../i18n'
  import { checkForUpdates, installUpdate, settings, updateSettings, updateState } from '../stores'

  const state = $derived($updateState)
  const version = $derived(state?.latest?.version ?? '')
  const busy = $derived(state?.status === 'checking' || state?.status === 'downloading')
  const percent = $derived(
    state && state.totalBytes > 0 ? Math.floor((state.downloadedBytes / state.totalBytes) * 100) : 0
  )
  const checkedAt = $derived(
    state?.checkedAt
      ? new Date(state.checkedAt).toLocaleString($locale ?? undefined, {
          dateStyle: 'medium',
          timeStyle: 'short'
        })
      : ''
  )
</script>

<div class="field">
  <p class="current">{$t('updates.current', { values: { version: state?.current ?? '' } })}</p>
  <p class="status" role="status" data-status={state?.status ?? 'idle'}>
    {$t(`updates.status.${state?.status ?? 'idle'}`, { values: { version, percent } })}
  </p>
  {#if state?.status === 'downloading'}
    <progress max="100" value={percent} aria-label={$t('updates.progress')}></progress>
  {/if}
  {#if state?.status === 'failed' && state.error}<pre class="error">{state.error}</pre>{/if}
  {#if checkedAt}<span class="hint">{$t('updates.lastCheck', { values: { date: checkedAt } })}</span
    >{/if}
  <div class="row">
    <button
      type="button"
      class="dlg-button"
      disabled={busy}
      onclick={() => checkForUpdates(bridge)}
    >
      {$t('updates.checkNow')}
    </button>
    {#if state?.status === 'ready'}
      <button type="button" class="dlg-button primary" onclick={() => installUpdate(bridge)}>
        {$t(state.installs ? 'updates.installRestart' : 'updates.showFile')}
      </button>
    {/if}
    {#if state?.latest}
      <button
        type="button"
        class="dlg-button plain"
        onclick={() => state?.latest && bridge.system.openUrl(state.latest.notesUrl)}
      >
        {$t('updates.whatsNew')}
      </button>
    {/if}
  </div>
</div>

<div class="field check">
  <label>
    <input
      type="checkbox"
      checked={$settings?.checkUpdates ?? true}
      aria-describedby="setting-updates-hint"
      onchange={(event) => updateSettings(bridge, { checkUpdates: event.currentTarget.checked })}
    />
    {$t('updates.auto')}
  </label>
  <span class="hint" id="setting-updates-hint">{$t('updates.autoHint')}</span>
</div>

<style>
  .current {
    margin: 0;
    font-weight: 700;
    font-size: 13.5px;
  }
  .status {
    margin: 4px 0 8px;
  }
  progress {
    width: 100%;
    height: 8px;
  }
  .error {
    margin: 0 0 8px;
    white-space: pre-wrap;
    font: 400 12px var(--mono);
    color: var(--err);
  }
  .row {
    display: flex;
    gap: 8px;
    margin-top: 8px;
    flex-wrap: wrap;
  }
  .check label {
    display: flex;
    gap: 8px;
    align-items: center;
    font-weight: 700;
    font-size: 13.5px;
  }
  .check .hint {
    padding-left: 26px;
  }
</style>
