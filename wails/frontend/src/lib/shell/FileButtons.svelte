<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { activeNeedsSave, newFile, openFileFromDialog, saveActiveFile } from '../stores'
</script>

<div class="file-buttons" role="group" aria-label={$t('file.menu')}>
  <button
    class="icon-btn"
    type="button"
    aria-label={$t('file.new')}
    title={$t('file.newTip')}
    onclick={() => void newFile(bridge)}
  >
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" />
      <path d="M14 3v5h5M12 11v6M9 14h6" />
    </svg>
  </button>
  <button
    class="icon-btn"
    type="button"
    aria-label={$t('file.open')}
    title={$t('file.openTip')}
    onclick={() => void openFileFromDialog(bridge)}
  >
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
    </svg>
  </button>
  <button
    class="icon-btn"
    type="button"
    aria-label={$t('file.save')}
    title={$t('file.saveTip')}
    disabled={!$activeNeedsSave}
    onclick={() => void saveActiveFile(bridge)}
  >
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M5 3h11l4 4v12a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z" />
      <path d="M7 3v5h8V3M7 21v-7h10v7" />
    </svg>
  </button>
</div>

<style>
  .file-buttons {
    display: flex;
    gap: 4px;
  }
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    padding: 0;
    border-radius: 8px;
    border: 1px solid var(--line);
    background: var(--win);
    color: var(--ink);
    cursor: pointer;
  }
  .icon-btn:hover:not(:disabled) {
    background: var(--rail);
  }
  .icon-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  svg {
    width: 18px;
    height: 18px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.8;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
</style>
