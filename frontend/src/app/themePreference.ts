import { THEME_PRESETS, type ThemeMode, type ThemePreset } from '../theme'

const THEME_KEY = 'gigmann.theme'
const THEME_PRESET_KEY = 'gigmann.theme.preset'

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

/** loadThemePreset resolves the saved colour preset, defaulting to Gigmann blue. */
export function loadThemePreset(): ThemePreset {
  const stored = localStorage.getItem(THEME_PRESET_KEY)
  return stored && stored in THEME_PRESETS ? (stored as ThemePreset) : 'gigmann'
}

/** saveThemePreset persists the selected colour preset. */
export function saveThemePreset(preset: ThemePreset): void {
  localStorage.setItem(THEME_PRESET_KEY, preset)
}
