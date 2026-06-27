import type { ThemeMode } from '../theme'

const THEME_KEY = 'gigmann.theme'

/**
 * loadThemeMode resolves the initial colour mode: an explicitly saved choice
 * wins, otherwise the operating-system preference (prefers-color-scheme), and
 * finally light. matchMedia is feature-detected so this is safe under jsdom/SSR.
 */
export function loadThemeMode(): ThemeMode {
  const stored = localStorage.getItem(THEME_KEY)
  if (stored === 'light' || stored === 'dark') {
    return stored
  }
  if (typeof window.matchMedia === 'function' && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }
  return 'light'
}

/** saveThemeMode persists the chosen colour mode so it survives a reload. */
export function saveThemeMode(mode: ThemeMode): void {
  localStorage.setItem(THEME_KEY, mode)
}
