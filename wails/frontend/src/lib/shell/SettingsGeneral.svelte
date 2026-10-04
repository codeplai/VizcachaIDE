<script lang="ts">
  import { bridge } from '../bridge'
  import type { CodeLanguage, LanguageSetting, ThemeSetting } from '../domain'
  import { t } from '../i18n'
  import { profiles, settings, updateSettings } from '../stores'

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
  <label for="setting-code-language">{$t('settings.defaultCodeLanguage')}</label>
  <select
    id="setting-code-language"
    value={$settings?.defaultCodeLanguage}
    onchange={(event) =>
      updateSettings(bridge, { defaultCodeLanguage: event.currentTarget.value as CodeLanguage })}
  >
    {#each $profiles as profile (profile.id)}
      <option value={profile.id}>{$t(profile.nameKey)}</option>
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
