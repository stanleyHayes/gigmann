import { describe, it, expect } from 'vitest'
import { buildTheme, statusColors, THEME_PRESETS } from './theme'

describe('buildTheme', () => {
  it('builds light and dark themes', () => {
    expect(buildTheme('light').palette.mode).toBe('light')
    expect(buildTheme('dark').palette.mode).toBe('dark')
  })

  it('exposes status colours for all severities', () => {
    expect(statusColors.good).toMatch(/^#/)
    expect(statusColors.watch).toMatch(/^#/)
    expect(statusColors.critical).toMatch(/^#/)
  })

  it('builds each accent preset', () => {
    for (const preset of Object.keys(THEME_PRESETS) as Array<keyof typeof THEME_PRESETS>) {
      expect(buildTheme('light', preset).palette.primary.main).toBe(THEME_PRESETS[preset].primary)
    }
  })

  it('derives action backgrounds from the selected preset', () => {
    const gigmann = buildTheme('light', 'gigmann')
    const cedar = buildTheme('light', 'cedar')

    expect(gigmann.palette.action.selected).not.toBe(cedar.palette.action.selected)
    expect(cedar.palette.action.hover).toContain('rgba')
  })
})
