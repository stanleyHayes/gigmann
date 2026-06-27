import { describe, expect, it } from 'vitest'

import { fmt } from './locale'
import { t } from './messages'

describe('i18n', () => {
  it('formats cedis with the GHS symbol', () => {
    expect(fmt.cedis(85000)).toMatch(/85,000/)
    expect(fmt.cedis(85000)).toMatch(/(GH₵|₵|GHS)/)
  })
  it('formats numbers and dates for en-GH', () => {
    expect(fmt.number(1234567)).toBe('1,234,567')
    expect(fmt.date('2026-06-09')).toMatch(/2026/)
  })
  it('looks up catalog messages', () => {
    expect(t('nav.today')).toBe('Today')
    expect(t('brief.source.local')).toMatch(/Deterministic/)
  })
})
