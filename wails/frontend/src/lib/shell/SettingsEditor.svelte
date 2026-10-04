<script lang="ts">
  import { bridge } from '../bridge'
  import { t } from '../i18n'
  import { MAX_FONT_SIZE, MIN_FONT_SIZE, clampFontSize, settings, updateSettings } from '../stores'
</script>

<div class="field">
  <label for="setting-font-size">{$t('settings.fontSize')}</label>
  <input
    id="setting-font-size"
    type="number"
    min={MIN_FONT_SIZE}
    max={MAX_FONT_SIZE}
    value={$settings?.fontSize}
    aria-describedby="setting-font-size-hint"
    onchange={(event) =>
      updateSettings(bridge, { fontSize: clampFontSize(event.currentTarget.valueAsNumber) })}
  />
  <span class="hint" id="setting-font-size-hint">{$t('settings.fontSizeHint')}</span>
</div>

<div class="field check">
  <label>
    <input
      type="checkbox"
      checked={$settings?.formatOnSave ?? true}
      aria-describedby="setting-format-hint"
      onchange={(event) => updateSettings(bridge, { formatOnSave: event.currentTarget.checked })}
    />
    {$t('settings.formatOnSave')}
  </label>
  <span class="hint" id="setting-format-hint">{$t('settings.formatOnSaveHint')}</span>
</div>

<div class="field check">
  <label>
    <input
      type="checkbox"
      checked={$settings?.inlayHints ?? true}
      onchange={(event) => updateSettings(bridge, { inlayHints: event.currentTarget.checked })}
    />
    {$t('settings.inlayHints')}
  </label>
</div>

<style>
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
