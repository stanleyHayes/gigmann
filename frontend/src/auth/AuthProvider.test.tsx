import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('../api/client', () => ({
  api: { GET: vi.fn(), POST: vi.fn() },
}))

import { api } from '../api/client'
import { AppProviders } from '../app/providers'
import { LoginScreen } from '../screens/LoginScreen'
import { AuthProvider } from './AuthProvider'
import { useAuth } from './authContext'
import { setToken } from './authStore'

const post = api.POST as unknown as ReturnType<typeof vi.fn>
const get = api.GET as unknown as ReturnType<typeof vi.fn>

function Probe() {
  const { isAuthenticated, user } = useAuth()
  return <div data-testid="probe">{isAuthenticated ? `auth:${user?.name ?? '...'}` : 'anon'}</div>
}

function renderAuth(node = <LoginScreen />) {
  return render(
    <AppProviders>
      <AuthProvider>
        {node}
        <Probe />
      </AuthProvider>
    </AppProviders>,
  )
}

describe('AuthProvider', () => {
  beforeEach(() => {
    setToken(null)
    post.mockReset()
    get.mockReset()
  })

  it('authenticates on a successful login', async () => {
    post.mockResolvedValue({ data: { token: 'tok', refresh_token: 'ref', user: { id: 'u1', name: 'Sammy Adjei', role: 'executive' } }, error: undefined })
    get.mockResolvedValue({ data: { id: 'u1', name: 'Sammy Adjei', role: 'executive' }, error: undefined })

    renderAuth()
    expect(screen.getByTestId('probe')).toHaveTextContent('anon')

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'ceo@gigmann.health' } })
    fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'ahenfie-demo' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(screen.getByTestId('probe')).toHaveTextContent('auth:Sammy Adjei'))
  })

  it('shows an error on bad credentials', async () => {
    post.mockResolvedValue({ data: undefined, error: { error: 'invalid_credentials' } })

    renderAuth()
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'ceo@gigmann.health' } })
    fireEvent.change(screen.getByLabelText(/^password$/i), { target: { value: 'wrong' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(screen.getByText(/invalid email or password/i)).toBeInTheDocument())
    expect(screen.getByTestId('probe')).toHaveTextContent('anon')
  })

  it('prompts for a code when MFA is required', async () => {
    post.mockResolvedValue({ data: undefined, error: { error: 'mfa_required' } })

    renderAuth()
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'ceo@gigmann.health' } })
    fireEvent.change(screen.getByLabelText(/^password/i), { target: { value: 'pw' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(screen.getByLabelText(/authenticator or recovery code/i)).toBeInTheDocument())
  })

  it('toggles password visibility on the sign-in form', () => {
    renderAuth()

    const password = screen.getByLabelText(/^password$/i)
    expect(password).toHaveAttribute('type', 'password')

    fireEvent.click(screen.getByRole('button', { name: /show sign-in password/i }))
    expect(password).toHaveAttribute('type', 'text')

    fireEvent.click(screen.getByRole('button', { name: /hide sign-in password/i }))
    expect(password).toHaveAttribute('type', 'password')
  })

  it('walks through the password reset flow', async () => {
    post.mockImplementation(async (path: string) => {
      if (path === '/api/v1/auth/password-reset/request') {
        return { data: { message: 'ready', reset_token: 'reset-token-123456' }, error: undefined }
      }
      if (path === '/api/v1/auth/password-reset/confirm') {
        return { data: undefined, error: undefined }
      }
      return { data: undefined, error: { error: 'unexpected' } }
    })

    renderAuth()
    fireEvent.click(screen.getByRole('button', { name: /forgot password/i }))
    fireEvent.click(screen.getByRole('button', { name: /send reset instructions/i }))

    await waitFor(() => expect(screen.getByLabelText(/reset token/i)).toHaveValue('reset-token-123456'))
    const newPassword = screen.getByLabelText(/^new password$/i)
    expect(newPassword).toHaveAttribute('type', 'password')
    fireEvent.click(screen.getByRole('button', { name: /show new password/i }))
    expect(newPassword).toHaveAttribute('type', 'text')
    fireEvent.change(newPassword, { target: { value: 'new-password' } })
    fireEvent.click(screen.getByRole('button', { name: /^reset password$/i }))

    await waitFor(() => expect(screen.getByText(/password reset/i)).toBeInTheDocument())
    expect(screen.getByRole('button', { name: /^sign in$/i })).toBeInTheDocument()
  })
})
