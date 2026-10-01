import type { ThemeSetting } from '../domain'

/** "system" removes the override so `prefers-color-scheme` decides; the others force a theme. */
export const applyTheme = (theme: ThemeSetting): void => {
  const root = document.documentElement
  if (theme === 'system') root.removeAttribute('data-theme')
  else root.setAttribute('data-theme', theme)
}
