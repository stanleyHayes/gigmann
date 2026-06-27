import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { loadThemeMode, saveThemeMode } from './themePreference'

describe('themePreference', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.unstubAllGlobals()
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('returns a saved choice over everything else', () => {
    saveThemeMode('dark')
    expect(loadThemeMode()).toBe('dark')
    saveThemeMode('light')
    expect(loadThemeMode()).toBe('light')
  })

  it('falls back to the OS preference when nothing is saved', () => {
    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({ matches: true }))
    expect(loadThemeMode()).toBe('dark')
  })

  it('defaults to light when nothing is saved and the OS prefers light', () => {
    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({ matches: false }))
    expect(loadThemeMode()).toBe('light')
  })

  it('defaults to light when matchMedia is unavailable (jsdom/SSR)', () => {
    // No matchMedia stub: jsdom does not implement it.
    expect(loadThemeMode()).toBe('light')
  })

  it('persists across reloads', () => {
    saveThemeMode('dark')
    expect(localStorage.getItem('gigmann.theme')).toBe('dark')
    // A fresh load (no stub) reads the stored value, not the OS default.
    expect(loadThemeMode()).toBe('dark')
  })
})
