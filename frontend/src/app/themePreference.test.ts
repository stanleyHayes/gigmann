import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { loadThemeMode, loadThemePreset, saveThemeMode, saveThemePreset } from './themePreference'

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

  it('persists a theme preset', () => {
    saveThemePreset('cedar')
    expect(loadThemePreset()).toBe('cedar')
    expect(localStorage.getItem('gigmann.theme.preset')).toBe('cedar')
  })

  it('falls back to the default preset for unknown stored values', () => {
    localStorage.setItem('gigmann.theme.preset', 'neon')
    expect(loadThemePreset()).toBe('gigmann')
  })
})
