<script lang="ts">
  import { bridge } from '../bridge'
  import type { LanguageSetting, ThemeSetting } from '../domain'
  import { t } from '../i18n'
  import { programArguments, settings, updateSettings } from '../stores'

  const languages: { id: LanguageSetting; label: string }[] = [
    { id: 'auto', label: 'settings.languageAuto' },
    { id: 'en', label: 'language.en' },
    { id: 'es', label: 'language.es' }
  ]
  const themes: { id: ThemeSetting; label: string }[] = [
    { id: 'system', label: 'settings.themeSystem' },
    { id: 'light', label: 'settings.themeLight' },
    { id: 'dark', label: 'settings.themeDark' }
  ]
</script>

<div class="field">
  <label for="setting-language">{$t('settings.language')}</label>
  <select
    id="setting-language"
    value={$settings?.language}
    onchange={(event) =>
      updateSettings(bridge, { language: event.currentTarget.value as LanguageSetting })}
  >
    {#each languages as option (option.id)}
      <option value={option.id}>{$t(option.label)}</option>
    {/each}
  </select>
</div>

<div class="field">
  <label for="setting-theme">{$t('settings.theme')}</label>
  <select
    id="setting-theme"
    value={$settings?.theme}
    onchange={(event) =>
      updateSettings(bridge, { theme: event.currentTarget.value as ThemeSetting })}
  >
    {#each themes as option (option.id)}
      <option value={option.id}>{$t(option.label)}</option>
    {/each}
  </select>
</div>

<div class="field">
  <label for="setting-program-args">{$t('settings.programArgs')}</label>
  <input
    id="setting-program-args"
    type="text"
    bind:value={$programArguments}
    aria-describedby="setting-program-args-hint"
  />
  <span class="hint" id="setting-program-args-hint">{$t('settings.programArgsHint')}</span>
</div>
