import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const h = vi.hoisted(() => ({
  login: vi.fn(),
  requestMutate: vi.fn(),
  confirmMutate: vi.fn(),
}))

vi.mock('../auth/authContext', () => ({
  useAuth: () => ({ login: h.login, loginPending: false, loginError: undefined, mfaRequired: false }),
}))
vi.mock('../api/usePasswordReset', () => ({
  usePasswordResetRequest: () => ({ mutate: h.requestMutate, isPending: false, isError: false, reset: vi.fn() }),
  usePasswordResetConfirm: () => ({ mutate: h.confirmMutate, isPending: false, isError: false, reset: vi.fn() }),
}))

import { LoginScreen } from './LoginScreen'

function renderLogin() {
  return render(<LoginScreen />, { wrapper: MemoryRouter })
}

describe('LoginScreen', () => {
  beforeEach(() => {
    h.login.mockReset()
    h.requestMutate.mockReset()
    h.confirmMutate.mockReset()
  })

  it('signs in with the entered credentials', () => {
    renderLogin()
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'ahenfie-demo' } })
    fireEvent.click(screen.getByRole('button', { name: /^sign in$/i }))
    expect(h.login).toHaveBeenCalledWith('ceo@gigmann.health', 'ahenfie-demo', undefined)
  })

  it('toggles password visibility', () => {
    renderLogin()
    fireEvent.click(screen.getByLabelText(/show sign-in password/i))
    expect(screen.getByLabelText(/hide sign-in password/i)).toBeInTheDocument()
  })

  it('switches to the reset-request flow and requests a token', () => {
    renderLogin()
    fireEvent.click(screen.getByRole('button', { name: /forgot password/i }))

    const emailField = screen.getByLabelText(/account email/i)
    fireEvent.change(emailField, { target: { value: 'ceo@gigmann.health' } })
    fireEvent.submit(emailField.closest('form') as HTMLFormElement)
    expect(h.requestMutate).toHaveBeenCalledWith('ceo@gigmann.health', expect.anything())
  })
})
