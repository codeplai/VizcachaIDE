import type { LanguageSetting } from './domain'

export type Language = 'en' | 'es'

/** "auto" follows the system: any Spanish locale (es, es-PE, ...) gives Spanish, the rest English. */
export const resolveLanguage = (setting: LanguageSetting, systemLanguage: string): Language => {
  if (setting === 'en' || setting === 'es') return setting
  return systemLanguage.toLowerCase().startsWith('es') ? 'es' : 'en'
}

export const systemLanguage = (): string =>
  typeof navigator === 'undefined' ? 'en' : (navigator.language ?? 'en')
