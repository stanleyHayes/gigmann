import { describe, it, expect } from 'vitest'
import { buildTheme, statusColors } from './theme'

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
})
