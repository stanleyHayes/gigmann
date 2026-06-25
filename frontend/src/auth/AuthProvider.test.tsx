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
    post.mockResolvedValue({ data: { token: 'tok', user: { id: 'u1', name: 'Sammy Adjei', role: 'executive' } }, error: undefined })
    get.mockResolvedValue({ data: { id: 'u1', name: 'Sammy Adjei', role: 'executive' }, error: undefined })

    renderAuth()
    expect(screen.getByTestId('probe')).toHaveTextContent('anon')

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'ceo@gigmann.health' } })
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'ahenfie-demo' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(screen.getByTestId('probe')).toHaveTextContent('auth:Sammy Adjei'))
  })

  it('shows an error on bad credentials', async () => {
    post.mockResolvedValue({ data: undefined, error: { error: 'invalid_credentials' } })

    renderAuth()
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'ceo@gigmann.health' } })
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'wrong' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(screen.getByText(/invalid email or password/i)).toBeInTheDocument())
    expect(screen.getByTestId('probe')).toHaveTextContent('anon')
  })
})
