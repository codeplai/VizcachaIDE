import { IntlMessageFormat } from 'intl-messageformat'
import { get } from 'svelte/store'
import { beforeAll, describe, expect, it } from 'vitest'
import { resolveLanguage } from '../language'
import { applyLanguage, formatSeconds, locale, setupI18n, t } from '.'
import en from './locales/en.json'
import es from './locales/es.json'

describe('catalogs', () => {
  it('have exactly the same keys in en and es', () => {
    expect(Object.keys(es).sort()).toEqual(Object.keys(en).sort())
  })

  it('are not empty and contain the UX copy keys', () => {
    expect(Object.keys(en).length).toBeGreaterThan(100)
    for (const key of ['actions.run', 'panels.output', 'empty.files', 'run.compileFailed']) {
      expect(en).toHaveProperty([key])
    }
  })

  it('parse as ICU messages in both languages', () => {
    for (const [lang, catalog] of [
      ['en', en],
      ['es', es]
    ] as const) {
      for (const [key, text] of Object.entries(catalog)) {
        expect(() => new IntlMessageFormat(text, lang), `${lang}:${key}`).not.toThrow()
      }
    }
  })

  it('keep the same placeholders in en and es', () => {
    const names = (text: string) => [...text.matchAll(/\{(\w+)[,}]/g)].map((m) => m[1]).sort()
    for (const key of Object.keys(en) as (keyof typeof en)[]) {
      expect(names(es[key]), key).toEqual(names(en[key]))
    }
  })
})

describe('language', () => {
  beforeAll(() => {
    setupI18n('en')
  })

  it('follows the system for "auto"', () => {
    expect(resolveLanguage('auto', 'es-PE')).toBe('es')
    expect(resolveLanguage('auto', 'es')).toBe('es')
    expect(resolveLanguage('auto', 'en-US')).toBe('en')
    expect(resolveLanguage('auto', 'fr-FR')).toBe('en')
    expect(resolveLanguage('en', 'es-PE')).toBe('en')
  })

  it('translates with exact UX copy and switches language', async () => {
    await applyLanguage('en')
    await Promise.resolve()
    expect(get(t)('actions.stopDebugging')).toBe('Stop debugging')
    applyLanguage('es')
    await Promise.resolve()
    expect(get(locale)).toBe('es')
    expect(get(t)('actions.stopDebugging')).toBe('Terminar depuración')
  })

  it('handles plurals', async () => {
    applyLanguage('en')
    await Promise.resolve()
    const text = (count: number) => get(t)('run.compileFailed', { values: { count } })
    expect(text(1)).toContain('found 1 problem.')
    expect(text(2)).toContain('found 2 problems.')
  })

  it('formats seconds by language', () => {
    expect(formatSeconds('en', 0.4)).toBe('0.4')
    expect(formatSeconds('es', 0.4)).toBe('0,4')
  })
})
