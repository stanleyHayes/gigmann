import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { SettingsScreen } from './SettingsScreen'

type Enroll = { mutate: ReturnType<typeof vi.fn>; isPending: boolean; data: { secret: string; otpauth_uri: string } | undefined }
type Confirm = { mutate: ReturnType<typeof vi.fn>; isPending: boolean; isError: boolean; isSuccess: boolean }
type Prefs = { data: { watched_metrics: string[]; thresholds: Record<string, number> } | undefined; isLoading: boolean }
type SavePrefs = { mutate: ReturnType<typeof vi.fn>; isPending: boolean; isSuccess: boolean }

const hoisted = vi.hoisted(() => ({
  enroll: { mutate: vi.fn(), isPending: false, data: undefined } as Enroll,
  confirm: { mutate: vi.fn(), isPending: false, isError: false, isSuccess: false } as Confirm,
  prefs: { data: { watched_metrics: [], thresholds: {} }, isLoading: false } as Prefs,
  savePrefs: { mutate: vi.fn(), isPending: false, isSuccess: false } as SavePrefs,
}))

vi.mock('../api/useMfa', () => ({
  useMfaEnroll: () => hoisted.enroll,
  useMfaConfirm: () => hoisted.confirm,
}))
vi.mock('../api/usePreferences', () => ({
  usePreferences: () => hoisted.prefs,
  useSavePreferences: () => hoisted.savePrefs,
}))

describe('SettingsScreen', () => {
  beforeEach(() => {
    hoisted.enroll = { mutate: vi.fn(), isPending: false, data: undefined }
    hoisted.confirm = { mutate: vi.fn(), isPending: false, isError: false, isSuccess: false }
    hoisted.prefs = { data: { watched_metrics: [], thresholds: {} }, isLoading: false }
    hoisted.savePrefs = { mutate: vi.fn(), isPending: false, isSuccess: false }
  })

  it('starts enrollment', () => {
    render(<SettingsScreen />)
    fireEvent.click(screen.getByRole('button', { name: /set up two-factor/i }))
    expect(hoisted.enroll.mutate).toHaveBeenCalled()
  })

  it('shows the secret and confirms a code', () => {
    hoisted.enroll = { mutate: vi.fn(), isPending: false, data: { secret: 'ABC234', otpauth_uri: 'otpauth://x' } }
    render(<SettingsScreen />)
    expect(screen.getByText('ABC234')).toBeInTheDocument()
    fireEvent.change(screen.getByLabelText(/authenticator code/i), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }))
    expect(hoisted.confirm.mutate).toHaveBeenCalledWith({ secret: 'ABC234', code: '123456' })
  })

  it('confirms success', () => {
    hoisted.confirm = { mutate: vi.fn(), isPending: false, isError: false, isSuccess: true }
    render(<SettingsScreen />)
    expect(screen.getByText(/two-factor authentication is on/i)).toBeInTheDocument()
  })

  it('saves watched-metric preferences', () => {
    hoisted.prefs = { data: { watched_metrics: ['revenue'], thresholds: {} }, isLoading: false }
    render(<SettingsScreen />)
    // pre-checked from preferences
    expect(screen.getByRole('checkbox', { name: /revenue/i })).toBeChecked()
    fireEvent.click(screen.getByRole('checkbox', { name: /occupancy/i }))
    fireEvent.click(screen.getByRole('button', { name: /save preferences/i }))
    expect(hoisted.savePrefs.mutate).toHaveBeenCalledWith({
      watched_metrics: ['revenue', 'occupancy'],
      thresholds: {},
    })
  })
})
