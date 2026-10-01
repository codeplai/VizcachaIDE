import { _, addMessages, init, locale } from 'svelte-i18n'
import type { LanguageSetting } from '../domain'
import { resolveLanguage, systemLanguage, type Language } from '../language'
import en from './locales/en.json'
import es from './locales/es.json'

/** Translate with `$t('actions.run')` or `$t('run.starting', { values: { file } })`. */
export const t = _
export { locale }

export const setupI18n = (setting: LanguageSetting = 'auto'): Language => {
  addMessages('en', en)
  addMessages('es', es)
  const initialLocale = resolveLanguage(setting, systemLanguage())
  void init({ fallbackLocale: 'en', initialLocale })
  return initialLocale
}

/** Applies the language chosen in Settings ("auto" follows the system). */
export const applyLanguage = (setting: LanguageSetting): Language => {
  const language = resolveLanguage(setting, systemLanguage())
  void locale.set(language)
  document.documentElement.lang = language
  return language
}

/** Formats a number of seconds like "0.4" or "0,4" for the current language. */
export const formatSeconds = (language: string, seconds: number): string =>
  new Intl.NumberFormat(language, { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(
    seconds
  )
