import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { SettingsScreen } from './SettingsScreen'

vi.mock('qrcode', () => ({
  default: {
    toDataURL: vi.fn().mockResolvedValue('data:image/png;base64,qr'),
  },
}))

type Enroll = { mutate: ReturnType<typeof vi.fn>; isPending: boolean; data: { secret: string; otpauth_uri: string } | undefined }
type Confirm = {
  mutate: ReturnType<typeof vi.fn>
  reset: ReturnType<typeof vi.fn>
  isPending: boolean
  isError: boolean
  isSuccess: boolean
  data: { recovery_codes: string[] } | undefined
}
type Disable = {
  mutate: ReturnType<typeof vi.fn>
  reset: ReturnType<typeof vi.fn>
  isPending: boolean
  isError: boolean
  isSuccess: boolean
}
type Prefs = { data: { watched_metrics: string[]; thresholds: Record<string, number> } | undefined; isLoading: boolean }
type SavePrefs = { mutate: ReturnType<typeof vi.fn>; isPending: boolean; isSuccess: boolean }

const hoisted = vi.hoisted(() => ({
  enroll: { mutate: vi.fn(), isPending: false, data: undefined } as Enroll,
  confirm: { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: false, data: undefined } as Confirm,
  disable: { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: false } as Disable,
  auth: { user: { id: 'u1', name: 'Sammy Adjei', role: 'executive' as const, mfa_enabled: false } },
  prefs: { data: { watched_metrics: [], thresholds: {} }, isLoading: false } as Prefs,
  savePrefs: { mutate: vi.fn(), isPending: false, isSuccess: false } as SavePrefs,
}))

vi.mock('../api/useMfa', () => ({
  useMfaEnroll: () => hoisted.enroll,
  useMfaConfirm: () => hoisted.confirm,
  useMfaDisable: () => hoisted.disable,
}))
vi.mock('../auth/authContext', () => ({
  useAuth: () => hoisted.auth,
}))
vi.mock('../api/usePreferences', () => ({
  usePreferences: () => hoisted.prefs,
  useSavePreferences: () => hoisted.savePrefs,
}))

describe('SettingsScreen', () => {
  beforeEach(() => {
    hoisted.enroll = { mutate: vi.fn(), isPending: false, data: undefined }
    hoisted.confirm = { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: false, data: undefined }
    hoisted.disable = { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: false }
    hoisted.auth = { user: { id: 'u1', name: 'Sammy Adjei', role: 'executive', mfa_enabled: false } }
    hoisted.prefs = { data: { watched_metrics: [], thresholds: {} }, isLoading: false }
    hoisted.savePrefs = { mutate: vi.fn(), isPending: false, isSuccess: false }
  })

  it('starts enrollment', () => {
    render(<SettingsScreen />)
    fireEvent.click(screen.getByRole('button', { name: /set up two-factor/i }))
    expect(hoisted.enroll.mutate).toHaveBeenCalled()
  })

  it('shows the QR code, secret, and confirms a code', async () => {
    hoisted.enroll = { mutate: vi.fn(), isPending: false, data: { secret: 'ABC234', otpauth_uri: 'otpauth://x' } }
    render(<SettingsScreen />)
    expect(screen.getByText('ABC234')).toBeInTheDocument()
    await waitFor(() =>
      expect(screen.getByRole('img', { name: /qr code/i })).toHaveAttribute('src', 'data:image/png;base64,qr'),
    )
    fireEvent.change(screen.getByLabelText(/authenticator code/i), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }))
    expect(hoisted.confirm.mutate).toHaveBeenCalledWith({ secret: 'ABC234', code: '123456' })
  })

  it('confirms success', () => {
    hoisted.confirm = {
      mutate: vi.fn(),
      reset: vi.fn(),
      isPending: false,
      isError: false,
      isSuccess: true,
      data: { recovery_codes: ['ABCD-1234-EFGH-5678', 'WXYZ-9876-IJKL-5432'] },
    }
    render(<SettingsScreen />)
    expect(screen.getByText(/two-factor authentication is on/i)).toBeInTheDocument()
    expect(screen.getByText('ABCD-1234-EFGH-5678')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /copy recovery codes/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /disable two-factor/i })).toBeInTheDocument()
  })

  it('disables MFA with a current code when already enabled', () => {
    hoisted.auth = { user: { id: 'u1', name: 'Sammy Adjei', role: 'executive', mfa_enabled: true } }
    render(<SettingsScreen />)
    fireEvent.change(screen.getByLabelText(/code to disable/i), { target: { value: 'ABCD-1234-EFGH-5678' } })
    fireEvent.click(screen.getByRole('button', { name: /disable two-factor/i }))
    expect(hoisted.disable.mutate).toHaveBeenCalledWith({ code: 'ABCD-1234-EFGH-5678' })
  })

  it('shows setup after MFA has been disabled', () => {
    hoisted.auth = { user: { id: 'u1', name: 'Sammy Adjei', role: 'executive', mfa_enabled: true } }
    hoisted.disable = { mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false, isSuccess: true }
    render(<SettingsScreen />)
    expect(screen.getByText(/two-factor authentication is off/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /set up two-factor/i })).toBeInTheDocument()
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
